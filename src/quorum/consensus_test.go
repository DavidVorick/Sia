package quorum

import (
	"common"
	"common/crypto"
	"reflect"
	"testing"
	//"time"
)

func TestHeartbeatEncoding(t *testing.T) {
	// encode a nil heartbeat
	var hb *heartbeat
	ehb, err := hb.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// create entropy for the heartbeat
	hb = new(heartbeat)
	entropy, err := crypto.RandomByteSlice(common.EntropyVolume)
	if err != nil {
		t.Fatal(err)
	}
	copy(hb.entropy[:], entropy)

	// create a public key
	pubKey, _, err := crypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// add each type of update to the map
	// currently there is only one type of update
	joinRequest := &JoinRequest{
		Sibling: Sibling{
			index:     255,
			address:   bootstrapAddress,
			publicKey: pubKey,
		},
	}
	hb.updates = make(map[Update]Update)
	hb.updates[joinRequest] = joinRequest

	// encode and decode the filled out heartbeat
	ehb, err = hb.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	// decode into a nil heartbeat
	var dhb *heartbeat
	err = dhb.GobDecode(ehb)
	if err == nil {
		t.Error("heartbeat.GobDecode accepts a nil heartbeat")
	}

	// decode into non-nil heartbeat
	dhb = new(heartbeat)
	err = dhb.GobDecode(ehb)
	if err != nil {
		t.Fatal(err)
	}

	// reflect.DeepEqual checks each value, including for the maps
	equal := reflect.DeepEqual(hb, dhb)
	if !equal {
		t.Error("heartbeat not identical after being encoded then decoded")
	}
}

func TestSignedHeartbeatEncoding(t *testing.T) {
	// tbi - the structure of a signed heartbeat is going to change, no point
	// in writing tests at the moment
}

func TestNewSignedHeartbeat(t *testing.T) {
	// tbi
}

/*
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
	var newSeed common.Entropy
	err = p1.processHeartbeat(sh0.heartbeat, &newSeed)
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
}*/
