package state

import (
	"erasure"
	"fmt"
	"io"
)

func RSEncode(input io.Reader, segments [QuorumSize]io.Writer, k int) (atoms uint16, err error) {
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
	if k < 1 || k >= int(QuorumSize) {
		err = fmt.Errorf("K must be between zero and %v (exclusive)", QuorumSize)
		return
	}
	m := int(QuorumSize) - k

	// read from the reader enough to build 1 atom on the quorum, then encode it
	// to a single atom, which is then written to all of the writers
	atom := make([]byte, AtomSize*int(k))
	var n int
	for n, err = input.Read(atom); err == nil || n > 0; atoms++ {
		if atoms == AtomsPerSector {
			err = fmt.Errorf("Exceeded max atoms per sector")
			return
		}

		var encodedSegments [][]byte
		encodedSegments, err = erasure.ReedSolomonEncode(k, m, atom)
		for i := range segments {
			segments[i].Write(encodedSegments[i])
		}
		n, err = input.Read(atom)
	}

	// check that at least 1 atom was created, and return
	if atoms == 0 {
		fmt.Println(err)
		err = fmt.Errorf("No data read from reader!")
	} else {
		err = nil
	}
	return
}

func RSRecover(segments []io.Reader, indicies []byte, output io.Writer, k int) (atoms uint16, err error) {
	if k < 1 || k >= int(QuorumSize) {
		err = fmt.Errorf("K must be between zero and %v, have %v", QuorumSize, k)
		return
	}

	if segments == nil {
		err = fmt.Errorf("Cannot recover from nil reader slice.")
		return
	}
	if len(segments) < k {
		err = fmt.Errorf("Insufficient input segments to recover sector.")
		return
	}
	for i := 0; i < k; i++ {
		if segments[i] == nil {
			err = fmt.Errorf("Reader %v is nil, cannot recover from a nil reader.", i)
			return
		}
	}
	if output == nil {
		err = fmt.Errorf("Cannot recover to nil io writer.")
		return
	}

	// inidicies gets error checked during call to erasure.Recover

	// create k atoms that are read into from segments
	atomsSlice := make([][]byte, k)
	for i := range atomsSlice {
		atomsSlice[i] = make([]byte, AtomSize)
	}

	// in a loop, read into atoms and call recover
	var n int
	var finished bool
	for {
		atoms++
		for i := range atomsSlice {
			n, err = segments[i].Read(atomsSlice[i])
			if err != nil && n == 0 {
				finished = true
			}
		}
		if finished {
			break
		}

		// got a bunch of new data, now recover it
		var recoveredAtom []byte
		recoveredAtom, err = erasure.ReedSolomonRecover(k, int(QuorumSize)-k, atomsSlice, indicies)
		if err != nil {
			return
		}
		output.Write(recoveredAtom)
	}

	if atoms == 0 {
		err = fmt.Errorf("Unable to read from all Readers")
	} else {
		err = nil
	}
	return
}
