package quorum

import (
	"erasure"
	"fmt"
	"io"
)

func RSEncode(input io.Reader, segments [QuorumSize]io.Writer, k byte) (atoms uint16, err error) {
	// check for nil inputs
	if input == nil {
		err = fmt.Errorf("Received nil input!")
		return
	}
	for i := range segments {
		if segments[i] == nil {
			err = fmt.Errorf("Received nil input within segments slice!")
			return
		}
	}

	// check k for sane value, then determine m
	if k < 1 || k >= byte(QuorumSize) {
		err = fmt.Errorf("K must be between zero and", QuorumSize, "(exclusive)")
		return
	}
	m := byte(QuorumSize) - k

	// read from the reader enough to build 1 atom on the quorum, then encode it
	// to a single atom, which is then written to all of the writers
	atom := make([]byte, AtomSize*int(k))
	for _, err = input.Read(atom); err != nil; atoms++ {
		if atoms == AtomsPerSector {
			err = fmt.Errorf("Exceeded max atoms per sector")
			return
		}

		var encodedSegments [][]byte
		encodedSegments, err = erasure.EncodeRedundancy(k, m, atom)
		for i := range segments {
			segments[i].Write(encodedSegments[i])
		}
		_, err = input.Read(atom)
	}

	// check that at least 1 atom was created, and return
	if atoms == 0 {
		err = fmt.Errorf("No data read from reader!")
	}
	return
}
