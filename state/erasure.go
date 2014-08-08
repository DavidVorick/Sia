package state

import (
	"errors"
	"fmt"
	"io"

	"github.com/NebulousLabs/Sia/erasure"
)

// RSEncode acts as a wrapper around erasure.ReedSolomonEncode for
// state-specific encoding operations. It is less flexible than
// erasure.ReedSolomonEncode, and returns the number of atoms per sector.
func RSEncode(input io.Reader, segments [QuorumSize]io.Writer, k int) (atoms uint16, err error) {
	// check for nil inputs
	if input == nil {
		err = errors.New("received nil input")
		return
	}
	for i := range segments {
		if segments[i] == nil {
			err = fmt.Errorf("segment %d is nil", i)
			return
		}
	}

	// check k for sane value, then determine m
	if k < 1 || k >= int(QuorumSize) {
		err = fmt.Errorf("k must be between zero and %v (exclusive)", QuorumSize)
		return
	}
	m := int(QuorumSize) - k

	// read from the reader enough to build 1 atom on the quorum, then encode it
	// to a single atom, which is then written to all of the writers
	atom := make([]byte, AtomSize*int(k))
	var readErr error
	for n, readErr := input.Read(atom); readErr == nil || n > 0; atoms++ {
		if atoms == AtomsPerSector {
			err = errors.New("exceeded max atoms per sector")
			return
		}

		var encodedSegments [][]byte
		encodedSegments, err = erasure.ReedSolomonEncode(k, m, atom)
		for i := range segments {
			segments[i].Write(encodedSegments[i])
		}
		n, readErr = input.Read(atom)
	}

	// check that at least 1 atom was created
	if atoms == 0 {
		fmt.Println(readErr) // remove?
		err = errors.New("no data read from reader")
	}
	return
}

// RSRecover acts as a wrapper around erasure.ReedSolomonRecover for
// state-specific encoding operations. It is less flexible than
// erasure.ReedSolomonRecover, and returns the number of atoms per sector.
func RSRecover(segments []io.Reader, indices []byte, output io.Writer, k int) (atoms uint16, err error) {
	if k < 1 || k >= int(QuorumSize) {
		err = fmt.Errorf("k must be between zero and %v", QuorumSize)
		return
	}

	if len(segments) < k {
		err = errors.New("insufficient input segments to recover sector")
		return
	}
	for i := 0; i < k; i++ {
		if segments[i] == nil {
			err = fmt.Errorf("Reader %v is nil", i)
			return
		}
	}
	if output == nil {
		err = errors.New("cannot write to nil output")
		return
	}

	// indices gets error-checked during call to erasure.Recover

	// create k atoms that are read into from segments
	atomsSlice := make([][]byte, k)
	for i := range atomsSlice {
		atomsSlice[i] = make([]byte, AtomSize)
	}

	// in a loop, read into atoms and call recover
loop:
	for {
		atoms++
		for i := range atomsSlice {
			n, _ := segments[i].Read(atomsSlice[i])
			if n != AtomSize {
				break loop
			}
		}

		// got a bunch of new data, now recover it
		var recoveredAtom []byte
		recoveredAtom, err = erasure.ReedSolomonRecover(k, int(QuorumSize)-k, atomsSlice, indices)
		if err != nil {
			return
		}
		output.Write(recoveredAtom)
	}

	if atoms == 0 {
		err = errors.New("unable to read from all Readers")
	}
	return
}
