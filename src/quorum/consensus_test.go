package quorum

import (
	"common"
	"common/crypto"
	"testing"
	"time"
)

// Verify that newHeartbeat() produces valid heartbeats
func TestNewHeartbeat(t *testing.T) {
	// tbi
}

func TestHeartbeatEncoding(t *testing.T) {
	// marshal an empty heartbeat
	hb := new(heartbeat)
	mhb, err := hb.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// unmarshal the empty heartbeat
	uhb := new(heartbeat)
	err = uhb.GobDecode(mhb)
	if err != nil {
		t.Fatal(err)
	}

	// test for equivalency
	if hb.entropy != uhb.entropy {
		t.Fatal("EntropyStage1 not identical upon umarshalling")
	}

	// test encoding with bad input
	err = uhb.GobDecode(nil)
	if err == nil {
		t.Error("able to decode a nil byte slice")
	}

	// fuzz over random potential values of heartbeat
}

func TestSignHeartbeat(t *testing.T) {
	// tbi
}

func TestSignedHeartbeatEncoding(t *testing.T) {
	// Test for bad inputs
	var bad *SignedHeartbeat
	bad = nil
	_, err := bad.GobEncode()
	if err == nil {
		t.Error("Should not encode a nil signedHeartbeat")
	}
	err = bad.GobDecode(nil)
	if err == nil {
		t.Error("Should not be able to decode a nil byte slice")
	}

	// Test the encoding and decoding of a simple signed heartbeat
	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	sh, err := p.newSignedHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	esh, err := sh.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	dsh := new(SignedHeartbeat)
	err = dsh.GobDecode(esh)
	if err != nil {
		t.Fatal(err)
	}

	// check encoding and decoding of a signedHeartbeat with many signatures
}

// Test takes .66 seconds to run... why?
func TestHandleSignedHeartbeat(t *testing.T) {
	// create a state and populate it with the signatories as siblings
	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// create keypairs
	pubKey1, secKey1, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, secKey2, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// create siblings and add them to s
	var p1 Sibling
	var p2 Sibling
	p.self.index = 0
	p1.index = 1
	p2.index = 2
	p1.publicKey = pubKey1
	p2.publicKey = pubKey2
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
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(esh)
	if err != nil {
		t.Fatal(err)
	}
	sh.signatures = make([]crypto.Signature, 2)
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
	p.heartbeats[sh.signatories[0]] = make(map[crypto.TruncatedHash]*heartbeat)

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
	sh.heartbeatHash, err = crypto.CalculateTruncatedHash(ehb)
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
	sh.signatures = make([]crypto.Signature, 2)
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
	p.currentStep = common.QuorumSize
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
	time.Sleep(common.StepDuration)
}

func TestTossSibling(t *testing.T) {
	// tbi
}

// Check that valid heartbeats are accepted and invalid heartbeats are rejected
func TestProcessHeartbeat(t *testing.T) {
	// create states and add them to each other
	p0, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	p1, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	p0.self.index = 0
	p1.self.index = 1
	p0.addNewSibling(p1.self)
	p1.addNewSibling(p0.self)

	// check that a valid heartbeat passes
	sh0, err := p0.newSignedHeartbeat()
	if err != nil {
		t.Fatal(err)
	}
	_, err = p1.quorum.processHeartbeat(sh0.heartbeat)
	if err != nil {
		t.Error("processHeartbeat threw out a valid heartbeat:", err)
	}
}

// TestCompile should probably be reviewed and rehashed
func TestCompile(t *testing.T) {
	// tbi
}

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// verify that tick is updating CurrentStep
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	p.stepLock.Lock()
	if p.currentStep != 2 {
		t.Fatal("s.currentStep failed to update correctly:", p.currentStep)
	}
	p.stepLock.Unlock()
}

// ensures Tick() calles compile() and then resets the counter to step 1
func TestCompilationTick(t *testing.T) {
	// test takes common.StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	// create state, set values for compile
	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}
	p.stepLock.Lock()
	p.currentStep = common.QuorumSize
	p.stepLock.Unlock()

	// verify that tick is wrapping around properly
	time.Sleep(common.StepDuration)
	time.Sleep(time.Second)
	p.stepLock.Lock()
	if p.currentStep != 1 {
		t.Error("p.currentStep failed to roll over:", p.currentStep)
	}
	p.stepLock.Unlock()
}
