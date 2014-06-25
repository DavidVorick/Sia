package client

import (
	"fmt"
	"io"
	"math"
	"network"
	"os"
	"participant"
	"quorum"
	"siacrypto"
)

func CalculateAtoms(filename string, k byte) (atoms int, err error) {
	multiplier := float64(k) / float64(quorum.QuorumSize)
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

func (c *Client) UploadFile(id quorum.WalletID, filename string, k byte) {
	var siblings [quorum.QuorumSize]*quorum.Sibling
	err := c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.Siblings",
		Args: struct{}{},
		Resp: &siblings,
	})
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}

	// take the file and produce a bunch of erasure coded atoms written one piece
	// at a time to be MerkleCollapsed and then uploaded to the siblings.
	// the approach is to create a bunch of files, one for each erasure coded section
	var writerSegments [quorum.QuorumSize]io.Writer
	var fileSegments [quorum.QuorumSize]*os.File
	for i := range fileSegments {
		tmpname := fmt.Sprintf("%s/%s.%v.tmp", os.TempDir(), filename, i)
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

	atomsWritten, err := quorum.RSEncode(file, writerSegments, k)
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}

	// resize the sector to exactly big enough

	// create the upload announcement
	var hashSet [quorum.QuorumSize]siacrypto.Hash
	for i := range fileSegments {
		_, err := fileSegments[i].Seek(0, 0)
		if err != nil {
			panic(err)
		}

		hashSet[i] = quorum.MerkleCollapse(fileSegments[i])
	}
	sectorHash := quorum.SectorHash(hashSet)

	// Now that the files have been written to 1 atom at a time, rewind them to
	// the beginning and create diffs for each file. Then upload the diffs to
	// each silbing via RPC
	currentSegment := make([]byte, int(atomsWritten)*quorum.AtomSize)
	for i := range fileSegments {
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
		c.router.SendMessage(&network.Message{
			Dest: siblings[i].Address(),
			Proc: "Participant.ReceieveDiff",
			Args: diff,
			Resp: nil,
		})
	}
}
