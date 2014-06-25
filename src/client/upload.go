package client

import (
	"fmt"
	"io"
	"math"
	"network"
	"os"
	"participant"
	"quorum"
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

	err = quorum.RSEncode(file, writerSegments, k)
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
		return
	}

	// Now that the files have been written to 1 atom at a time, rewind them to
	// the beginning and create diffs for each file. Then upload the diffs to
	// each silbing via RPC
}
