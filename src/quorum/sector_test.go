package quorum

import (
	"bytes"
	"siacrypto"
	"testing"
)

func TestMerkleCollapse(t *testing.T) {
	q := new(Quorum)
	for i := 0; i < 33; i++ {
		randomBytes := siacrypto.RandomByteSlice(i * AtomSize)
		b := bytes.NewBuffer(randomBytes)
		q.MerkleCollapse(b)
	}

	if testing.Short() {
		t.Skip()
	}

	for i := 0; i < 12; i++ {
		numAtoms, _ := siacrypto.RandomInt(1024)
		randomBytes := siacrypto.RandomByteSlice(numAtoms * AtomSize)
		b := bytes.NewBuffer(randomBytes)
		q.MerkleCollapse(b)
	}
}
