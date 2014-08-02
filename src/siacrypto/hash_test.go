package siacrypto

import (
	"testing"
)

// TestHashBytes tests that the computed hash of "foo" matches its reference hash.
func TestHashBytes(t *testing.T) {
	referenceHash := Hash{
		247, 251, 186, 110, 6, 54, 248, 144, 229, 111, 187, 243, 40, 62, 82, 76,
		111, 163, 32, 74, 226, 152, 56, 45, 98, 71, 65, 208, 220, 102, 56, 50,
	} // in bytes, the truncated hash of string "foo"

	// compute hash and compare to reference
	hash := HashBytes([]byte("foo"))
	if hash != referenceHash {
		t.Fatal("Hash produced unexpected value")
	}

	// calculate nil hashes, should not cause a panic
	hash = HashBytes(nil)
}

// TestHashObject tests that HashObject hashes multiple distinct types without error.
func TestHashObject(t *testing.T) {
	objs := []interface{}{
		nil,
		struct{}{},
		true,
		1,
		3.14,
		"foo",
		[]byte{1, 2, 3},
		Hash{},
	}
	for obj := range objs {
		_, err := HashObject(obj)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkHashBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashBytes(RandomByteSlice(1024))
	}
}

func BenchmarkHashObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashObject(RandomByteSlice(1024))
	}
}
