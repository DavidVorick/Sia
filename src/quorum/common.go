package quorum

const (
	// How big a single segment of data is for a host, in bytes
	AtomsPerStack  int = 16777216 // 64GB - each atom is 4kb
	MinSegmentSize int = 512
	MaxSegmentSize int = 1048576 // 1 MB
)
