package siacrypto

import (
	"testing"
)

func TestPublicKeyCompare(t *testing.T) {
	// compare nil public keys
	var pk0 *PublicKey
	var pk1 *PublicKey
	compare := pk0.Compare(pk1)
	if compare {
		t.Error("Comparing nil public keys return true")
	}

	// compare when one public key is nil
	pk0, _, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = pk0.Compare(pk1)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}
	compare = pk1.Compare(pk0)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}

	// compare unequal public keys
	pk1, _, err = CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = pk0.Compare(pk1)
	if compare {
		t.Error("Arbitray public keys being compared as identical")
	}
	compare = pk1.Compare(pk0)
	if compare {
		t.Error("Arbitrary public keys being compared as identical")
	}

	// compare a key to itself
	compare = pk0.Compare(pk0)
	if !compare {
		t.Error("A key returns false when comparing with itself")
	}

	// compare some manufactured identical keys
	// compare when nil values are contained within the struct (lower priority)
}

func TestPublicKeyEncoding(t *testing.T) {
	// Encode and Decode nil values
	var pk *PublicKey
	_, _ = pk.GobEncode() // checking for panics
	pk = new(PublicKey)
	_, _ = pk.GobEncode() // checking for panics

	_ = pk.GobDecode(nil) // checking for panics

	// Encode and then Decode, see if the results are identical
	pubKey, _, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	ePubKey, err := pubKey.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = pk.GobDecode(ePubKey)
	if err != nil {
		t.Fatal(err)
	}
	compare := pk.Compare(pubKey)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
	}
	compare = pubKey.Compare(pk)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
	}

	// Decode bad values and wrong values
}
