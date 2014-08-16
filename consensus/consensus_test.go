package consensus

import (
	"testing"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
)

// TestConsensus is the catch-all function for testing the components of the
// Sia backend. Proper testing often requires a quorum (and even a full quorum)
// to be established and in full consensus. This takes many lines of code,
// which are all handled below.
func TestConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Create a keypair for the tether wallet.
	tetherWalletPK, tetherWalletSK, err := siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Create a new quorum.
	mr, err := network.NewRPCServer(11000)
	if err != nil {
		t.Fatal(err)
	}
	p, err := CreateBootstrapParticipant(mr, siafiles.TempFilename("TestConsensus-Start"), 1, tetherWalletPK)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the next update has been accepted.
	p.updatesLock.RLock()
	if len(p.updates[0]) != 1 {
		t.Error("Update for the next block has not been accepted by the participant.")
	}
	p.updatesLock.RUnlock()

	// Create a participant that will join the current existing
	// participant. No timing shortcuts can be taken here, as the new
	// participant is doing explicitly timed sleeping. The tether wallet
	// used will be the same wallet used to tether the bootstrap
	// participant.
	p.engineLock.RLock()
	quorumSiblings := p.engine.Metadata().Siblings
	p.engineLock.RUnlock()
	var quorumSiblingAddresses []network.Address
	for _, sibling := range quorumSiblings {
		if sibling.Inactive() {
			continue
		}
		quorumSiblingAddresses = append(quorumSiblingAddresses, sibling.Address)
	}
	joiningParticipant, err := CreateJoiningParticipant(mr, siafiles.TempFilename("TestConsensus-Join1"), 1, tetherWalletSK, quorumSiblingAddresses)
	if err != nil {
		t.Fatal(err)
	}

	// CreateJoiningParticipant won't return until it has fully integrated.
	// Test that the integration was successful.
	p.engineLock.RLock()
	if p.engine.Metadata().Siblings[0].Inactive() || p.engine.Metadata().Siblings[1].Inactive() {
		t.Error("Initial participant is not recognizing both siblings as active.")
	}
	p.engineLock.RUnlock()

	joiningParticipant.engineLock.RLock()
	if joiningParticipant.engine.Metadata().Siblings[0].Inactive() || joiningParticipant.engine.Metadata().Siblings[1].Inactive() {
		t.Error("Joined participant is not recognizing both siblings as active.")
	}
	joiningParticipant.engineLock.RUnlock()

	// Add 2 more participants simultaneously and see if everything is
	// stable upon completion. The mutexing is so that non-parallel
	// functions can run in parallel, while the program still has to wait
	// for both to finish.
	joinChan := make(chan *Participant)
	go func() {
		p, err := CreateJoiningParticipant(mr, siafiles.TempFilename("TestConsensus-Join2"), 1, tetherWalletSK, quorumSiblingAddresses)
		if err != nil {
			t.Fatal(err)
		}
		joinChan <- p
	}()
	go func() {
		p, err := CreateJoiningParticipant(mr, siafiles.TempFilename("TestConsensus-Join3"), 1, tetherWalletSK, quorumSiblingAddresses)
		if err != nil {
			t.Fatal(err)
		}
		joinChan <- p
	}()
	join2, join3 := <-joinChan, <-joinChan

	// At this point, there should be a full quorum, where each participant
	// recognized all other participants. We run a check to see that each
	// participant recognizes each other participant as active siblings.
	for i, participant := range []*Participant{p, joiningParticipant, join2, join3} {
		participant.engineLock.RLock()
		for j, sibling := range participant.engine.Metadata().Siblings {
			if sibling.Inactive() {
				t.Error("Sibling recognized as inactive for iterators", i, ",", j)
			}
		}
		participant.engineLock.RUnlock()
	}
}

/*
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
	sh.heartbeatHash = siacrypto.HashBytes(hbb)
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
		sh.heartbeatHash, err = siacrypto.HashBytes(ehb)
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
		time.Sleep(StepDuration)
}*/
