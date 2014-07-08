package consensus

import (
	"network"
	"os"
	"quorum"
	"siacrypto"
	"testing"
)

// create a participant with a quorum and then generate a new signed heartbeat,
// do basic checking to make sure there are no panics and no errors.
func TestNewSignedHeartbeat(t *testing.T) {
	p := new(Participant)
	p.self = new(quorum.Sibling)
	p.heartbeats[p.self.Index()] = make(map[siacrypto.Hash]*heartbeat)
	_, key, err := siacrypto.CreateKeyPair()
	p.secretKey = key
	if err != nil {
		t.Fatal(err)
	}

	err = p.newSignedHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.heartbeats[p.self.Index()]) == 0 {
		t.Error("a heartbeat was not added to the local list of heartbeats")
	}
}

// Test takes .66 seconds to run... try to get below .1
//
// TestHandleSignedHeartbeat checks for every type of possible malicious
// behavior and makes sure that all malicious heartbeats are detected and
// thrown out.
func TestHandleSignedHeartbeat(t *testing.T) {
	p := new(Participant)
	for i := range p.heartbeats {
		p.heartbeats[i] = make(map[siacrypto.Hash]*heartbeat)
	}
	p.messageRouter = new(network.DebugNetwork)
	p.quorum = *new(quorum.Quorum)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wd = wd + "/../../participantStorage/TestHandleSignedHeartbeat."
	p.quorum.SetWalletPrefix(wd)

	// create keypairs
	pubKey, secKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pubKey1, secKey1, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, secKey2, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create dummy wallet
	p.quorum.CreateBootstrapWallet(1, quorum.NewBalance(0, 15000), nil)
	wallet := p.quorum.LoadWallet(1)

	// create siblings and add them to the quorum
	sibling := quorum.NewSibling(BootstrapAddress, pubKey)
	sibling1 := quorum.NewSibling(BootstrapAddress, pubKey1)
	sibling2 := quorum.NewSibling(BootstrapAddress, pubKey2)
	p.quorum.AddSibling(wallet, sibling1)
	p.quorum.AddSibling(wallet, sibling2)

	// populate participant with a self and a secret key
	p.self = sibling
	p.secretKey = secKey

	// create SignedHeartbeat
	hb := new(heartbeat)
	sh := new(SignedHeartbeat)
	sh.heartbeat = hb
	hbb, _ := hb.GobEncode()
	sh.heartbeatHash = siacrypto.CalculateHash(hbb)
	sh.signatories = make([]byte, 2)
	sh.signatures = make([]siacrypto.Signature, 2)
	sh.signatories[0] = sibling1.Index()
	sh.signatories[1] = sibling2.Index()
	sig1, err := secKey1.Sign(sh.heartbeatHash[:])
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures[0] = sig1.Signature
	encSm, err := sig1.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := secKey2.Sign(encSm)
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures[1] = sig2.Signature

	// handle the signed heartbeat, expecting nil error
	err = p.HandleSignedHeartbeat(*sh, nil)
	if err != nil {
		t.Error(err)
	}

	/*
		// handle the signed heartbeat, expecting nil error
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != nil {
			t.Fatal(err)
		}

		// verify that a repeat heartbeat gets ignored
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != hsherrHaveHeartbeat {
			t.Error("expected heartbeat to get ignored as a duplicate:", err)
		}

		// save the signature from the old heartbeat to falsify the new heartbeat
		badSig := sh.signatures[1]

		// create a different heartbeat, this will be used to test the fail conditions
		sh, err = p.newSignedHeartbeat()
		if err != nil {
			t.Fatal(err)
		}
		ehb, err := sh.heartbeat.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		sh.heartbeatHash, err = siacrypto.CalculateHash(ehb)
		if err != nil {
			t.Fatal(err)
		}

		// verify a heartbeat with bad signatures is rejected
		sh.signatures[0] = badSig
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != hsherrInvalidSignature {
			t.Fatal("expected heartbeat to get ignored as having invalid signatures: ", err)
		}

		// verify that a non-sibling gets rejected
		sh.signatories[0] = 3
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != hsherrNonSibling {
			t.Error("expected non-sibling to be rejected: ", err)
		}

		// give heartbeat repeat signatures
		signature1, err = secKey1.Sign(sh.heartbeatHash[:])
		if err != nil {
			t.Fatal(err)
		}

		encSm, err = signature1.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		signature2, err = secKey1.Sign(encSm)
		if err != nil {
			t.Error(err)
		}

		// adjust signatories slice
		sh.signatures = make([]siacrypto.Signature, 2)
		sh.signatories = make([]byte, 2)
		sh.signatures[0] = signature1.Signature
		sh.signatures[1] = signature2.Signature
		sh.signatories[0] = 1
		sh.signatories[1] = 1

		// verify repeated signatures are rejected
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != hsherrDoubleSigned {
			t.Error("expected heartbeat to be rejected for duplicate signatures: ", err)
		}

		// remove second signature
		sh.signatures = sh.signatures[:1]
		sh.signatories = sh.signatories[:1]

		// handle heartbeat when tick is larger than num signatures
		p.stepLock.Lock()
		p.currentStep = 2
		p.stepLock.Unlock()
		err = p.HandleSignedHeartbeat(*sh, nil)
		if err != hsherrNoSync {
			t.Error("expected heartbeat to be rejected as out-of-sync: ", err)
		}

		// remaining tests require sleep
		if testing.Short() {
			t.Skip()
		}

		// send a heartbeat right at the edge of a new block
		p.stepLock.Lock()
		p.currentStep = QuorumSize
		p.stepLock.Unlock()

		// submit heartbeat in separate thread
		go func() {
			err = p.HandleSignedHeartbeat(*sh, nil)
			if err != nil {
				t.Fatal("expected heartbeat to succeed!: ", err)
			}
			// need some way to verify with the test that the funcion gets here
		}()

		p.stepLock.Lock()
		p.currentStep = 1
		p.stepLock.Unlock()
		time.Sleep(time.Second)
		time.Sleep(StepDuration)*/
}
