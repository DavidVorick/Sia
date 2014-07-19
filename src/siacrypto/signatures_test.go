package siacrypto

import (
	"testing"
)

// Test takes .4 seconds to run, which is too long
func TestSigning(t *testing.T) {
	// Create a keypair
	publicKey, secretKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// sign a nil message
	_, err = secretKey.Sign(nil)
	if err == nil {
		t.Error("Signed a nil message!")
	}

	// sign an empty message
	empty := make([]byte, 0)
	sig, err := secretKey.Sign(empty)
	if err != nil {
		t.Error("Error returned when signing an empty message")
	}

	// verify the empty message
	verified := publicKey.Verify(sig, empty)
	if !verified {
		t.Error("Signed empty message did not verify!")
	}

	if testing.Short() {
		t.Skip()
	}

	// verify empty message when signature is bad
	sig[0] = ^sig[0] // flip the bits on the first byte to guarantee corruption
	verified = publicKey.Verify(sig, empty)
	if verified {
		t.Error("Verified a signed empty message with forged signature")
	}

	// sign the message
	randomMessage := RandomByteSlice(100)
	sig, err = secretKey.Sign(randomMessage)
	if err != nil {
		return
	}

	// verify the signature
	verification := publicKey.Verify(sig, randomMessage)
	if !verification {
		t.Error("failed to verify a valid message")
	}

	// verify an imposter signature
	sig[0] = ^sig[0]
	verification = publicKey.Verify(sig, randomMessage)
	if verification {
		t.Error("sucessfully verified an invalid message")
	}

	// restore the signature and fake a message
	sig[0] = ^sig[0]
	randomMessage[0] = ^randomMessage[0]
	verification = publicKey.Verify(sig, randomMessage)
	if verification {
		t.Error("successfully verified a corrupted message")
	}
}
