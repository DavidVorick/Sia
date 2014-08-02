package siacrypto

import (
	"testing"
)

// TestSigning tests various cryptographic signing functions.
func TestSigning(t *testing.T) {
	// create a keypair
	publicKey, secretKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal("Failed to create keypair:", err)
	}

	// attempt to sign a nil message
	_, err = secretKey.Sign(nil)
	if err == nil {
		t.Error("Signed a nil message")
	}

	// sign an empty message
	empty := make([]byte, 0)
	sig, err := secretKey.Sign(empty)
	if err != nil {
		t.Error("Failed to sign an empty message:", err)
	}

	// verify the empty message
	verified := publicKey.Verify(sig, empty)
	if !verified {
		t.Error("Failed to verify signed empty message")
	}

	if testing.Short() {
		t.Skip()
	}

	// attempt to verify empty message with bad signature
	sig[0] = ^sig[0] // flip the bits on the first byte to guarantee corruption
	verified = publicKey.Verify(sig, empty)
	if verified {
		t.Error("Verified a signed empty message with bad signature")
	}

	// create and sign a random message
	randomMessage := RandomByteSlice(100)
	sig, err = secretKey.Sign(randomMessage)
	if err != nil {
		t.Fatal("Failed to sign message:", err)
	}

	// verify the signature
	verification := publicKey.Verify(sig, randomMessage)
	if !verification {
		t.Error("Failed to verify a valid signature")
	}

	// attempt to verify bad signature
	sig[0] = ^sig[0]
	verification = publicKey.Verify(sig, randomMessage)
	if verification {
		t.Error("Verified a message with bad signature")
	}

	// restore the signature, but use a different message
	sig[0] = ^sig[0]
	randomMessage[0] = ^randomMessage[0]
	verification = publicKey.Verify(sig, randomMessage)
	if verification {
		t.Error("Verified a different message with the same signature")
	}
}
