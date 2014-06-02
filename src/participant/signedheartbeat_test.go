package participant

/*import (
	"network"
	"quorum"
	"siacrypto"
	"testing"
	"time"
)

// Test takes .66 seconds to run... try to get below .1
func TestHandleSignedHeartbeat(t *testing.T) {
	p := new(Participant)
	p.quorum := new(quorum.Quorum)

	// create keypairs
	pubKey1, secKey1, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, secKey2, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create siblings and add them to s
	p1 := Sibling {
		index: 1,
		publicKey: pubKey1,
	}
	p2 := Sibling {
		index: 2,
		publicKey: pubKey2,
	}

	err = p.addNewSibling(p.self)
	if err != nil {
		t.Fatal(err)
	}
	err = p.addNewSibling(&p1)
	if err != nil {
		t.Fatal(err)
	}
	err = p.addNewSibling(&p2)
	if err != nil {
		t.Fatal(err)
	}

	// create SignedHeartbeat
	sh, err := p.newSignedHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	esh, err := sh.heartbeat.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	sh.heartbeatHash, err = siacrypto.CalculateTruncatedHash(esh)
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures = make([]siacrypto.Signature, 2)
	sh.signatories = make([]byte, 2)

	// Create a set of signatures for the SignedHeartbeat
	signature1, err := secKey1.Sign(sh.heartbeatHash[:])
	if err != nil {
		t.Fatal(err)
	}

	combinedMessage, err := signature1.CombinedMessage()
	if err != nil {
		t.Fatal(err)
	}
	signature2, err := secKey2.Sign(combinedMessage)
	if err != nil {
		t.Fatal(err)
	}

	// build a valid SignedHeartbeat
	sh.signatures[0] = signature1.Signature
	sh.signatures[1] = signature2.Signature
	sh.signatories[0] = 1
	sh.signatories[1] = 2

	// delete existing heartbeat from state; makes the remaining tests easier
	p.heartbeats[sh.signatories[0]] = make(map[siacrypto.TruncatedHash]*heartbeat)

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
	sh.heartbeatHash, err = siacrypto.CalculateTruncatedHash(ehb)
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

	combinedMessage, err = signature1.CombinedMessage()
	if err != nil {
		t.Fatal(err)
	}
	signature2, err = secKey1.Sign(combinedMessage)
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
	time.Sleep(StepDuration)
}*/
