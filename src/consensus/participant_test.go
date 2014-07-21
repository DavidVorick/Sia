package consensus

import (
	"network"
	"siacrypto"
	"testing"
)

// TestNewParticipnat runs NewParticipant and checks to see that all the basic
// items have been initialized.
func TestNewParticipant(t *testing.T) {
	// Test calling NewParticipant with a nil message router.
	p, err := NewParticipant(nil, "../../filesCreatedDuringTesting/TestNewParticipant")
	if err == nil {
		t.Error("Able to create a participant with a nil message router.")
	}

	mr, err := network.NewRPCServer(11200)
	if err != nil {
		t.Fatal(err)
	}
	p, err = NewParticipant(mr, "../../filesCreatedDuringTesting/TestNewParticipant")
	if err != nil {
		t.Fatal(err)
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

	// Test that the address has been initialized, and the handler has been
	// registered. ???
}

// TestBroadcast... don't really know how to write this test.
