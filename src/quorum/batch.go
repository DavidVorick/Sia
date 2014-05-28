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
