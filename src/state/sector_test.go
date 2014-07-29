package state

import (
	"bytes"
	"siacrypto"
	"testing"
)

func joinHash(h1, h2 siacrypto.Hash) siacrypto.Hash {
	return siacrypto.CalculateHash(append(h1[:], h2[:]...))
}

func TestMerkleEmpty(t *testing.T) {
	b := make([]byte, 7*AtomSize)
	merkleHash := MerkleCollapse(bytes.NewReader(b))

	// manually construct Merkle hash
	empty := make([]byte, AtomSize)
	l1 := siacrypto.CalculateHash(empty)
	l2 := joinHash(l1, l1)
	l3 := joinHash(l2, l2)

	r1 := siacrypto.CalculateHash(empty)
	r2 := joinHash(r1, r1)
	r3 := joinHash(r2, r1)

	finalHash := joinHash(l3, r3)

	if merkleHash != finalHash {
		t.Fatal("MerkleCollapse produced incorrect hash")
	}
}

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
	// generate random data
	var numAtoms uint16 = 7
	data := bytes.NewReader(siacrypto.RandomByteSlice(int(numAtoms) * AtomSize))

	var proofIndex uint16 = 0
	proofBase, proofStack := buildProof(data, numAtoms, proofIndex)

	// no need to call VerifyStorageProof directly; just simulate it
	data.Seek(0, 0)
	expectedHash := MerkleCollapse(data)
	initialHash := siacrypto.CalculateHash(proofBase)
	finalHash := foldHashes(initialHash, proofIndex, proofStack)

	if finalHash != expectedHash {
		t.Fatal("proof verification failed: hashes do not match")
	}

	// run foldHashes without enough proofs
	data.Seek(0, 0)
	finalHash = foldHashes(initialHash, proofIndex, proofStack[0:1])

	if finalHash == expectedHash {
		t.Fatal("invalid proof was verified")
	}

	if testing.Short() {
		t.Skip()
	}

	// ensure functions work for any tree configuration
	for i := uint16(1); i < 33; i++ {
		data = bytes.NewReader(siacrypto.RandomByteSlice(int(i) * AtomSize))
		proofIndex = siacrypto.RandomUint16() % i
		proofBase, proofStack = buildProof(data, i, proofIndex)

		data.Seek(0, 0)
		expectedHash = MerkleCollapse(data)
		initialHash = siacrypto.CalculateHash(proofBase)
		finalHash = foldHashes(initialHash, proofIndex, proofStack)

		if finalHash != expectedHash {
			t.Fatal("proof verification failed: hashes do not match", i, proofIndex)
		}
	}
}
