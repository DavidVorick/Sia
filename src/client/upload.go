package client

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"network"
	"os"
	"participant"
	"quorum"
	"quorum/script"
	"siacrypto"
	"siaencoding"
	"siafiles"
	"time"
)

func CalculateAtoms(filename string, k byte) (atoms int, err error) {
	multiplier := 1 / float64(k)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return
	}
	size := info.Size()

	floatAtoms := multiplier * float64(size) / float64(quorum.AtomSize)
	atoms = int(math.Ceil(floatAtoms))
	return
}

func calculatePadding(file *os.File, k byte) (padding uint32, err error) {
	info, err := file.Stat()
	if err != nil {
		return
	}
	size := 4 + info.Size()

	padding = uint32(size % int64(quorum.AtomSize) * int64(k))
	return
}

func (c *Client) UploadFile(id quorum.WalletID, filename string, k byte) {
	if c.genericWallets[id] == nil {
		fmt.Printf("Do not have access to wallet %v!\n", id)
		return
	}

	// Get a fresh list of siblings to have highest probability of success
	c.RetrieveSiblings()

	// take the file and produce a bunch of erasure coded atoms written one piece
	// at a time to be MerkleCollapsed and then uploaded to the siblings.
	// the approach is to create a bunch of files, one for each erasure coded section
	var writerSegments [quorum.QuorumSize]io.Writer
	var fileSegments [quorum.QuorumSize]*os.File
	nonce := siafiles.SafeFilename(siacrypto.RandomByteSlice(2))
	for i := range fileSegments {
		tmpname := fmt.Sprintf("%s/%s.%v.tmp", os.TempDir(), nonce, i)
		file, err := os.Create(tmpname)
		if err != nil {
			fmt.Printf("Upload: Error: %v\n", err)
			return
		}
		defer func(file *os.File, tmpname string) {
			file.Close()
			os.Remove(tmpname)
		}(file, tmpname)
		writerSegments[i] = file
		fileSegments[i] = file
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}
	defer file.Close()

	// When uploading in the generic case, some information is needed about
	// padding This gets put into it's own reader to prepend the encoding process
	paddingNeeded, err := calculatePadding(file, k)
	if err != nil {
		fmt.Println(err)
		return
	}
	encodedPadding := siaencoding.EncUint32(paddingNeeded)
	paddingBuffer := bytes.NewBuffer(encodedPadding)
	paddingAndFile := siafiles.NewDoubleReader(paddingBuffer, file)

	atomsWritten, err := quorum.RSEncode(paddingAndFile, writerSegments, k)
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}

	// resize the sector to exactly big enough
	// get address from the first non-nil participant
	input := script.ResizeSectorEraseInput(atomsWritten+1, k)
	input, err = script.SignInput(c.genericWallets[id].SK, input)
	if err != nil {
		return
	}
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: id,
			Input:    input,
		},
		Resp: nil,
	})
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}

	time.Sleep(time.Duration(quorum.QuorumSize) * participant.StepDuration)

	// figure out the hash now that the sector has been resized
	emptySegment := make([]byte, quorum.AtomSize*int(atomsWritten))
	b := bytes.NewBuffer(emptySegment)
	zeroMerkle := quorum.MerkleCollapse(b)
	emptyAtom := make([]byte, quorum.AtomSize)
	for i := 0; i < quorum.QuorumSize; i++ {
		copy(emptyAtom[i*siacrypto.HashSize:], zeroMerkle[:])
	}
	parentHash := siacrypto.CalculateHash(emptyAtom)

	// fetch the current block to determine a reasonable deadline
	deadline := quorum.MaxDeadline // cheating right now... will implement rest of deadline soon

	// get the hash set and the set for propose upload
	var hashSet [quorum.QuorumSize]siacrypto.Hash
	for i := range fileSegments {
		_, err := fileSegments[i].Seek(0, 0)
		if err != nil {
			panic(err)
		}

		hashSet[i] = quorum.MerkleCollapse(fileSegments[i])
	}
	sectorHash := quorum.SectorHash(hashSet)

	// Create and submit a propose upload script
	ua := quorum.UploadArgs{
		ParentHash:    parentHash,
		NewHashSet:    hashSet,
		AtomsChanged:  atomsWritten,
		Confirmations: k,
		Deadline:      deadline,
	}
	encUA, err := ua.GobEncode()
	if err != nil {
		panic(err)
	}
	si, err := script.SignInput(c.genericWallets[id].SK, script.ProposeUploadInput(encUA))
	if err != nil {
		panic(err)
	}
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: id,
			Input:    si,
		},
		Resp: nil,
	})

	// give enough time for the propose upload to complete
	time.Sleep(2 * time.Duration(quorum.QuorumSize) * participant.StepDuration)

	// Now that the files have been written to 1 atom at a time, rewind them to
	// the beginning and create diffs for each file. Then upload the diffs to
	// each silbing via RPC
	currentSegment := make([]byte, int(atomsWritten)*quorum.AtomSize)
	for i := range fileSegments {
		if c.siblings[i] == nil {
			continue
		}

		// read the appropriate segment into memory to be sent over RPC
		_, err = fileSegments[i].Seek(0, 0)
		if err != nil {
			panic(err)
		}
		_, err = fileSegments[i].Read(currentSegment)
		if err != nil {
			panic(err)
		}
		conversion := make([]participant.Conversion, 1)
		conversion[0].Offset = 0
		conversion[0].Delta = currentSegment
		diff := participant.UploadDiff{
			ID:            id,
			Hash:          sectorHash,
			ConversionSet: conversion,
		}

		// send the diff over RPC
		err = c.router.SendMessage(&network.Message{
			Dest: c.siblings[i].Address(),
			Proc: "Participant.ReceieveDiff",
			Args: diff,
			Resp: nil,
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}
