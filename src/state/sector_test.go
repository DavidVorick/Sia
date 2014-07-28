package state

import (
	"bytes"
	"siacrypto"
	"testing"
)

// TestMerkleCollapse tests that MerkleCollapse runs without error.
// It will be updated later to check that a known byte slice collapses
// to the current root-level hash.
func TestMerkleCollapse(t *testing.T) {
	for i := 0; i < 33; i++ {
		randomBytes := siacrypto.RandomByteSlice(i * AtomSize)
		b := bytes.NewBuffer(randomBytes)
		MerkleCollapse(b)
	}

	if testing.Short() {
		t.Skip()
	}

	for i := 0; i < 12; i++ {
		numAtoms, _ := siacrypto.RandomInt(1024)
		randomBytes := siacrypto.RandomByteSlice(numAtoms * AtomSize)
		b := bytes.NewBuffer(randomBytes)
		MerkleCollapse(b)
	}
}

// TestStorageProof tests the BuildStorageProof and VerifyStorageProof functions.
// It generates a storage using from random data, and verifies that the proof is correct.
func TestStorageProof(t *testing.T) {
	// catch panics, since proper error handling is not implemented yet
	defer func() {
		if p := recover(); p != nil {
			t.Fatal("caught panic:", p)
		}
	}()

	// generate random data
	var numAtoms uint16 = 16
	data := bytes.NewReader(siacrypto.RandomByteSlice(int(numAtoms) * AtomSize))

	// select random index for storage proof
	proofIndex := siacrypto.RandomUint16() % numAtoms

	proofBase, proofStack := buildProof(data, numAtoms, proofIndex)

	// no need to call VerifyStorageProof directly; just simulate it
	data.Seek(0, 0)
	expectedHash := MerkleCollapse(data)
	initialHash := siacrypto.CalculateHash(proofBase)
	finalHash := foldHashes(initialHash, proofIndex, 1, proofStack)

	if finalHash != expectedHash {
		t.Fatal("proof verification failed: hashes do not match")
	}
}
