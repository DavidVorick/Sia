package quorum

import (
	"os"
)

// a BID, or Batch ID, is the global logical address of a batch on Sia. A BID
// has no relationship to where on disk or in the AA tree the batch is stored.
// To perform lookups from a BID to the disk location of a BID, a batchMap must
// be used.
type CID [32]byte // not exactly sure what BID will end up looking like

// A cylinder is the set of 128 corresponding batches in a quorum.
type cylinder struct {
	ringPairs int
	ringMList []int
	ringAtoms []int
	cid       CID
}

// A batch is a group of sectors that is error-corrected together, and is the
// smallest unit that will move around on the network.
type batch struct {
	bid           *CID
	file          *os.File
	node          *batchNode
	sectorLengths []int
}

// A batchMap maps BIDs to their batch object
type batchMap map[CID]batch
