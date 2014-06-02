package quorum

import (
	"network"
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
	p1.publicKey = pubKey
	p0.publicKey = new(siacrypto.PublicKey)
	*p0.publicKey = *p1.publicKey

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
	p1.address.Port = 9987
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
	p1.publicKey = pubKey
	compare = p0.Compare(p1)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
	compare = p1.Compare(p0)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
}

func TestSiblingEncoding(t *testing.T) {
	// Try nil values
	var p *Sibling
	_, err := p.GobEncode()
	if err == nil {
		t.Error("Encoded nil sibling without error")
	}
	p = new(Sibling)
	_, err = p.GobEncode()
	if err == nil {
		t.Fatal("Should not be able to encode nil values")
	}

	// Make a bootstrap participant
	pubKey, _, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p.publicKey = pubKey
	p.address = network.Address{
		ID:   3,
		Host: "localhost",
		Port: 9950,
	}

	up := new(Sibling)
	ep, err := p.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = up.GobDecode(ep)
	if err != nil {
		t.Fatal(err)
	}

	if up.address != p.address {
		t.Error("up.address != p.address")
	}

	compare := up.publicKey.Compare(p.publicKey)
	if compare != true {
		t.Error("up.PublicKey != p.PublicKey")
	}

	// try to decode into nil participant
	up = nil
	err = up.GobDecode(ep)
	if err != nil {
		t.Error("falid to deceode into nil participant")
	}
}
