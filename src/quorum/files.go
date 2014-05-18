package quorum

import (
	"os"
)

// a BID, or Batch ID, is the global logical address of a batch on Sia. A BID
// has no relationship to where on disk or in the AA tree the batch is stored.
// To perform lookups from a BID to the disk location of a BID, a batchMap must
// be used.
type BID [32]byte

// A batch is a group of sectors that is error-corrected together, and is the
// smallest unit that will move around on the network.
type batch struct {
	file          *os.File
	node          *batchNode
	sectors       int
	sectorLengths []int
}

// A batchMap maps BIDs to their batch object
type batchMap map[BID]batch

// batchNode is the basic element in the AA tree that enables efficient
// selection of random sectors for verification
type batchNode struct {
	parent *batchNode
	left   *batchNode
	right  *batchNode

	leftWeight  int
	rightWeight int
	data        *batch
}
