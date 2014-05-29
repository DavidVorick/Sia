package quorum

import (
	"common"
	"siacrypto"
	"testing"
)

func TestSiblingCompare(t *testing.T) {
	var p0 *Sibling
	var p1 *Sibling

	// compare nil values
	compare := p0.compare(p1)
	if compare == true {
		t.Error("Comparing any nil participant should return false")
	}

	// compare when one is nil
	p0 = new(Sibling)
	compare = p0.compare(p1)
	if compare == true {
		t.Error("Comparing a zero participant to a nil participant should return false")
	}
	compare = p1.compare(p0)
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
	compare = p0.compare(p1)
	if !compare {
		t.Error("Comparing two zero participants should return true")
	}
	compare = p1.compare(p0)
	if !compare {
		t.Error("Comparing two zero participants should return true")
	}

	// compare when address are not equal
	p1.address.Port = 9987
	compare = p0.compare(p1)
	if compare {
		t.Error("Comparing two participants with different addresses should return false")
	}
	compare = p1.compare(p0)
	if compare {
		t.Error("Comparing two zero participants with different addresses should return false")
	}

	// compare when public keys are not equivalent
	pubKey, _, err = siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	p1.publicKey = pubKey
	compare = p0.compare(p1)
	if compare == true {
		t.Error("Comparing two participants with different public keys should return false")
	}
	compare = p1.compare(p0)
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
	p.address = bootstrapAddress

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

// check general case, check corner cases, and then do some fuzzing
func TestRandInt(t *testing.T) {
	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// check that it works in the vanilla case
	previousSeed := p.quorum.seed
	randInt, err := p.quorum.randInt(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if randInt < 0 || randInt >= 5 {
		t.Fatal("randInt returned but is not between the bounds")
	}

	// check that s.CurrentEntropy flipped to next value
	if previousSeed == p.quorum.seed {
		t.Error(previousSeed)
		t.Error(p.quorum.seed)
		t.Fatal("When calling randInt, s.CurrentEntropy was not changed")
	}

	// check the zero value
	randInt, err = p.quorum.randInt(0, 0)
	if err == nil {
		t.Fatal("Randint(0,0) should return a bounds error")
	}

	// fuzzing, skip for short tests
	if testing.Short() {
		t.Skip()
	}

	low := 0
	high := common.QuorumSize
	for i := 0; i < 100000; i++ {
		randInt, err = p.quorum.randInt(low, high)
		if err != nil {
			t.Fatal("randInt fuzzing error: ", err, " low: ", low, " high: ", high)
		}

		if randInt < low || randInt >= high {
			t.Fatal("randInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
}
