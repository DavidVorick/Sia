package consensus

import (
	"testing"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
)

// TestNewParticipnat runs NewParticipant and checks to see that all the basic
// items have been initialized.
func TestNewParticipant(t *testing.T) {
	// Test calling NewParticipant with a nil message router.
	p, err := newParticipant(nil, siafiles.TempFilename("TestNewParticipant"))
	if err == nil {
		t.Error("Created a participant with a nil message router")
	}

	mr, err := network.NewRPCServer(11200)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	p, err = newParticipant(mr, siafiles.TempFilename("TestNewParticipant"))
	if err != nil {
		t.Fatal("Failed to create participant:", err)
	}

	// Test that a keypair exists.
	var emptyPublicKey siacrypto.PublicKey
	var emptySecretKey siacrypto.SecretKey
	if p.publicKey == emptyPublicKey {
		t.Error("Public key not properly initialized")
	}
	if p.secretKey == emptySecretKey {
		t.Error("Secret key not properly initialized")
	}

	// Test that the siblingIndex has been set to the non-sibling value.
	if p.siblingIndex != ^byte(0) {
		t.Error("siblingIndex not initialized to ^byte(0)")
	}

	// Test that the address has been initialized, and the participant is
	// reachable.
	err = mr.Ping(p.address)
	if err != nil {
		t.Error("Participant not reachable:", err)
	}
}

// TestBroadcast... don't really know how to write this test.
