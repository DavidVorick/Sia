package quorum

import (
	"bytes"
	"erasure"
	"fmt"
)

// EncodeRing takes a Sector and encodes it as a Ring: a set of QuorumSize Segments that include redundancy.
// The encoding parameters are stored in params.
// k is the number of non-redundant segments, and b is the size of each segment. b is calculated from k.
// The erasure-coding algorithm requires that the original data must be k*b in size, so it is padded here as needed.
//
// The return value is a Ring.
// The first k Segments of the Ring are the original data split up.
// The remaining Segments are newly generated redundant data.
func EncodeRing(sec *Sector, params *EncodingParams) (ring [QuorumSize]Segment, err error) {
	k, b, length := params.GetValues()

	// check for legal size of k
	if k <= 0 || k >= QuorumSize {
		err = fmt.Errorf("k must be greater than 0 and smaller than %v", QuorumSize)
		return
	}

	// check for legal size of b
	if b < MinSegmentSize || b > MaxSegmentSize {
		err = fmt.Errorf("b must be greater than %v and smaller than %v", MinSegmentSize, MaxSegmentSize)
		return
	}

	// check for legal size of length
	if length != len(sec.Data) {
		err = fmt.Errorf("length mismatch: sector length %v != parameter length %v", len(sec.Data), length)
		return
	} else if length > MaxSegmentSize*QuorumSize {
		err = fmt.Errorf("length must be smaller than %v", MaxSegmentSize*QuorumSize)
	}

	// pad data as needed
	padding := k*b - len(sec.Data)
	paddedData := append(sec.Data, bytes.Repeat([]byte{0x00}, padding)...)

	// call the encoding function
	m := QuorumSize - k
	encodedData, err := erasure.EncodeRedundancy(k, m, paddedData)
	if err != nil {
		return
	}

	// copy data into ring
	for i := 0; i < QuorumSize; i++ {
		ring[i] = Segment{
			encodedData[i],
			uint8(i),
		}
	}

	return
}

// RebuildSector takes a Ring and returns a Sector containing the original data.
// The encoding parameters are stored in params.
// k must be equal to the number of non-redundant segments when the file was originally built.
// Because recovery is just a bunch of matrix operations, there is no way to tell if the data has been corrupted
// or if an incorrect value of k has been chosen. This error checking must happen before calling RebuildSector.
// Each Segment's Data must have the correct Index from when it was encoded.
func RebuildSector(ring []Segment, params *EncodingParams) (sec *Sector, err error) {
	k, b, length := params.GetValues()
	if k == 0 && b == 0 {
		err = fmt.Errorf("could not rebuild using uninitialized encoding parameters")
		return
	}

	// check for legal size of k
	if k > QuorumSize || k < 1 {
		err = fmt.Errorf("k must be greater than 0 but smaller than %v", QuorumSize)
		return
	}

	// check for legal size of b
	if b < MinSegmentSize || b > MaxSegmentSize {
		err = fmt.Errorf("b must be greater than %v and smaller than %v", MinSegmentSize, MaxSegmentSize)
		return
	}

	// check for legal size of length
	if length > MaxSegmentSize*QuorumSize {
		err = fmt.Errorf("length must be smaller than %v", MaxSegmentSize*QuorumSize)
	}

	// check for correct number of segments
	if len(ring) < k {
		err = fmt.Errorf("insufficient segments: expected at least %v, got %v", k, len(ring))
		return
	}

	// move all data into a single slice
	segmentData := make([][]byte, k)
	segmentIndicies := make([]int, k)
	for i := 0; i < k; i++ {
		if len(ring[i].Data) != b {
			err = fmt.Errorf("at least 1 Segment's Data field is the wrong length")
			return
		}

		segmentData[i] = ring[i].Data
		segmentIndicies[i] = int(ring[i].Index)
	}

	// call the recovery function
	recovered, err := erasure.Recover(k, QuorumSize-k, segmentData, segmentIndicies)
	if err != nil {
		return
	}

	// remove padding introduced by EncodeRing()
	sec, err = NewSector(recovered[:length])
	return
}
