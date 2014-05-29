package quorum

import (
	"os"
)

// A batch is a group of sectors that is error-corrected together, and is the
// smallest unit that will move around on the network.
type batch struct {
	bid           *CID
	file          *os.File
	node          *cylinderNode
	sectorLengths []int
}

// outline of a batch:
// 8 atoms for encoding the layered portion
// N atoms in the layered portion
// M sectors in the non-layered portion
