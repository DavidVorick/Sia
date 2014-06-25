package client

import (
	"fmt"
	"math"
	"network"
	"os"
	"participant"
	"quorum"
)

func CalculateAtoms(filename string, m byte) (atoms int, err error) {
	multiplier := float64(m) / float64(quorum.QuorumSize)
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

func (c *Client) UploadFile(id quorum.WalletID, filename string, m byte) {
	var siblings [quorum.QuorumSize]*quorum.Sibling
	err := c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.Siblings",
		Args: struct{}{},
		Resp: &siblings,
	})
	if err != nil {
		fmt.Printf("Upload: Error: %v\n", err)
	}

	// take the file and produce a bunch of erasure coded atoms written one piece
	// at a time to be MerkleCollapsed and then uploaded to the siblings.
}
