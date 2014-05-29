package siacrypto

import (
	"math/big"
	"testing"
)

// Basic testing of key creation, signing, and verification
// Implicitly tests SignedMessage.CombinedMessage()
//
// Test takes .4 seconds to run... why?
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
	msg, err := secretKey.Sign(empty)
	if err != nil {
		t.Error("Error returned when signing an empty message")
	}

	// verify the empty message
	verified := publicKey.Verify(&msg)
	if !verified {
		t.Error("Signed empty message did not verify!")
	}

	// verify empty message when signature is bad
	msg.Signature.R.Sub(msg.Signature.R, big.NewInt(1))
	verified = publicKey.Verify(&msg)
	if verified {
		t.Error("Verified a signed empty message with forged signature")
	}

	// sign using a nil key
	var nilKey *SecretKey
	_, err = nilKey.Sign(empty)
	if err == nil {
		t.Error("Signed with a nil key!")
	}

	// verify a nil signature
	verified = publicKey.Verify(nil)
	if verified {
		t.Error("Verified a nil signature...")
	}

	// create arbitrary message
	randomMessage, err := RandomByteSlice(20)
	if err != nil {
		t.Fatal(err)
	}

	// sign the message
	signedMessage, err := secretKey.Sign(randomMessage)
	if err != nil {
		return
	}

	// verify the signature
	verification := publicKey.Verify(&signedMessage)
	if !verification {
		t.Error("failed to verify a valid message")
	}

	// verify an imposter signature
	signedMessage.Signature.R.Sub(msg.Signature.R, big.NewInt(1))
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("sucessfully verified an invalid message")
	}

	// restore the signature and fake a message
	signedMessage.Signature.R.Add(msg.Signature.R, big.NewInt(1))
	signedMessage.Message[0] = 0
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("successfully verified a corrupted message")
	}
}

// There should probably be some benchmarking
