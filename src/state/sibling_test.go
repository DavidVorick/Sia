package state

import (
	"siacrypto"
	"testing"
)

func TestSiblingCompare(t *testing.T) {
	var p0 *Sibling
	var p1 *Sibling

	// compare nil values
	compare := p0.Compare(p1)
	if compare == true {
		t.Error("Comparing any nil participant should return false")
	}

	// compare when one is nil
	p0 = new(Sibling)
	compare = p0.Compare(p1)
	if compare == true {
		t.Error("Comparing a zero participant to a nil participant should return false")
	}
	compare = p1.Compare(p0)
	if compare == true {
		t.Error("Comparing a zero participant to a nil participant should return false")
	}

	// initialize each participant with a public key
	p1 = new(Sibling)
	pubKey, _, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p1.PublicKey = pubKey
	p0.PublicKey = new(siacrypto.PublicKey)
	*p0.PublicKey = *p1.PublicKey

	// compare initialized participants
	compare = p0.Compare(p1)
	if !compare {
		t.Error("Comparing two zero participants should return true")
	}
	compare = p1.Compare(p0)
	if !compare {
		t.Error("Comparing two zero participants should return true")
	}

	// compare when address are not equal
	p1.Address.Port = 9987
	compare = p0.Compare(p1)
	if compare {
		t.Error("Comparing two participants with different addresses should return false")
	}
	compare = p1.Compare(p0)
	if compare {
		t.Error("Comparing two zero participants with different addresses should return false")
	}

	// compare when public keys are not equivalent
	pubKey, _, err = siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p1.PublicKey = pubKey
	compare = p0.Compare(p1)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
	compare = p1.Compare(p0)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
}
