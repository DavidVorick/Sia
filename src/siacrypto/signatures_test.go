package siacrypto

import (
	"testing"
)

func TestPublicKeyCompare(t *testing.T) {
	// compare unequal public keys
	pk1, _, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	var pk0 PublicKey
	compare := pk0.Compare(pk1)
	if compare {
		t.Error("Arbitray public keys being compared as identical")
	}
	compare = pk1.Compare(pk0)
	if compare {
		t.Error("Arbitrary public keys being compared as identical")
	}

	// compare a key to itself
	compare = pk0.Compare(pk0)
	if !compare {
		t.Error("A key returns false when comparing with itself")
	}
}

func TestSecretKeyCompare(t *testing.T) {
	// Compare unequal secret keys.
	_, sk1, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	var sk0 SecretKey
	compare := sk0.Compare(sk1)
	if compare {
		t.Error("Arbitray public keys being compared as identical")
	}
	compare = sk1.Compare(sk0)
	if compare {
		t.Error("Arbitrary public keys being compared as identical")
	}

	// compare a key to itself
	compare = sk0.Compare(sk0)
	if !compare {
		t.Error("A key returns false when comparing with itself")
	}
}

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
	msg, err := secretKey.Sign(empty)
	if err != nil {
		t.Error("Error returned when signing an empty message")
	}

	// verify the empty message
	verified := publicKey.Verify(msg)
	if !verified {
		t.Error("Signed empty message did not verify!")
	}

	if testing.Short() {
		t.Skip()
	}

	// verify empty message when signature is bad
	msg.Signature[0] = ^msg.Signature[0] // flip the bits on the first byte to guarantee corruption
	verified = publicKey.Verify(msg)
	if verified {
		t.Error("Verified a signed empty message with forged signature")
	}

	// sign using a nil key
	var nilKey *SecretKey
	_, err = nilKey.Sign(empty)
	if err == nil {
		t.Error("Signed with a nil key!")
	}

	// sign the message
	signedMessage, err := secretKey.Sign(RandomByteSlice(20))
	if err != nil {
		return
	}

	// verify the signature
	verification := publicKey.Verify(signedMessage)
	if !verification {
		t.Error("failed to verify a valid message")
	}

	// verify an imposter signature
	signedMessage.Signature[0] = ^signedMessage.Signature[0]
	verification = publicKey.Verify(signedMessage)
	if verification {
		t.Error("sucessfully verified an invalid message")
	}

	// restore the signature and fake a message
	signedMessage.Signature[0] = ^signedMessage.Signature[0]
	signedMessage.Message[0] = ^signedMessage.Message[0]
	verification = publicKey.Verify(signedMessage)
	if verification {
		t.Error("successfully verified a corrupted message")
	}
}
