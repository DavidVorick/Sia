package erasure

import (
	"quorum"
	"siacrypto"
	"testing"
)

// Basic test for reed-solomon coding, verifies that standard input
// will produce the correct results.
func TestCoding(t *testing.T) {
	// set encoding parameters
	k := quorum.QuorumSize / 2
	m := quorum.QuorumSize - k
	b := 1024

	// create sector data
	randomBytes, err := siacrypto.RandomByteSlice(b * k)
	if err != nil {
		t.Fatal(err)
	}

	// create sector
	sec, err := quorum.NewSector(randomBytes)
	if err != nil {
		t.Fatal(err)
	}

	// calculate encoding parameters
	params := sec.CalculateParams(k)

	// encode data into a Ring
	ring, err := EncodeRing(sec, params)
	if err != nil {
		t.Fatal(err)
	}

	// create Ring from subset of encoded segments
	var newRing []quorum.Segment
	for i := m; i < quorum.QuorumSize; i++ {
		newRing = append(newRing, ring[i])
	}

	// recover original data
	newSec, err := RebuildSector(newRing, params)
	if err != nil {
		t.Fatal(err)
	}

	// compare to hash of data when first generated
	recoveredDataHash, err := siacrypto.CalculateHash(newSec.Data)
	if err != nil {
		t.Fatal(err)
	} else if recoveredDataHash != sec.Hash {
		t.Fatal("recovered data is different from original data")
	}

	// In every test, we check that the hashes equal
	// every other hash that gets created. This makes
	// me uneasy.
}

// At some point, there should be a long test that explores all of the edge cases.

// There should be a fuzzing test that explores random inputs. In particular, I would
// like to fuzz the 'RebuildSector' function

// There should also be a benchmarking test here.
