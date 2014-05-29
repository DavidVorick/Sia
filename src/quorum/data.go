package quorum

import (
	"common"
	"network"
	"siacrypto"
)

// A Sector is a logical block of data.
type Sector struct {
	Data []byte
	Hash siacrypto.Hash
}

// NewSector creates a Sector from data.
func NewSector(data []byte) (s *Sector, err error) {
	// calculate hash
	hash, err := siacrypto.CalculateHash(data)

	s = &Sector{data, hash}
	return
}

// A Segment is an erasure-coded piece of a Sector, containing both a subset of the original data and its corresponding index.
// A set of QuorumSize Segments forms a Ring.
type Segment struct {
	Data  []byte
	Index uint8
}

// A RingHeader contains all the metadata necessary to retrieve and rebuild a Sector from a Ring.
// This includes the hosts on which Ring Segments are stored, the encoding parameters, the hashes of each Segment.
type RingHeader struct {
	Hosts     [common.QuorumSize]network.Address
	Params    *EncodingParams
	SegHashes [common.QuorumSize]siacrypto.Hash
}

// EncodingParams are the parameters needed to perform erasure encoding and decoding.
// k is the number of non-redundant segments, and b is the number of bytes per segment.
// The length is also stored, because the encoding process may introduce padding.
type EncodingParams struct {
	k, b, length int
}

// CalculateParams creates a set of encoding parameters given a Sector and a k value.
func (s *Sector) CalculateParams(k int) *EncodingParams {
	// calculate length
	length := len(s.Data)

	// calculate b
	b := length / k
	if b%64 != 0 {
		b += 64 - (b % 64) // round up to nearest multiple of 64
	}

	return &EncodingParams{k, b, length}
}

func (e *EncodingParams) GetValues() (int, int, int) {
	return e.k, e.b, e.length
}
