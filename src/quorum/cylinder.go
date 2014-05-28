package quorum

// a CID, or Cylinder ID, is the global logical address of a batch on Sia. A CID
// has no relationship to where on disk or in the AA tree the batch is stored.
// To perform lookups from a CID to the disk location of a CID, a map musst be
// used.
type CID [32]byte // not exactly sure what CID will end up looking like

// A cylinder is the set of 128 corresponding batches in a quorum.
type cylinder struct {
	ringPairs int
	ringMList []int
	ringAtoms []int
	cid       CID
}

// A cylinderMap maps CIDs to their cylinder object within the cylinderTree
type cylinderMap map[CID]batch