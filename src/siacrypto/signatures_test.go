package siacrypto

import (
	"math/big"
	"testing"
)

func TestPublicKeyCompare(t *testing.T) {
	// compare nil public keys
	var pk0 *PublicKey
	var pk1 *PublicKey
	compare := pk0.Compare(pk1)
	if compare {
		t.Error("Comparing nil public keys return true")
	}

	// compare when one public key is nil
	pk0, _, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = pk0.Compare(pk1)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}
	compare = pk1.Compare(pk0)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}

	// compare unequal public keys
	pk1, _, err = CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = pk0.Compare(pk1)
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

	// compare some manufactured identical keys
	// compare when nil values are contained within the struct (lower priority)
}

func TestPublicKeyEncoding(t *testing.T) {
	// Encode and Decode nil values
	var pk *PublicKey
	_, _ = pk.GobEncode() // checking for panics
	pk = new(PublicKey)
	_, _ = pk.GobEncode() // checking for panics

	_ = pk.GobDecode(nil) // checking for panics

	// Encode and then Decode, see if the results are identical
	pubKey, _, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	ePubKey, err := pubKey.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = pk.GobDecode(ePubKey)
	if err != nil {
		t.Fatal(err)
	}
	compare := pk.Compare(pubKey)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
	}
	compare = pubKey.Compare(pk)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
	}

	// Decode bad values and wrong values
}

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
	msg.Signature.r.Sub(msg.Signature.r, big.NewInt(1))
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
	signedMessage.Signature.r.Sub(msg.Signature.r, big.NewInt(1))
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("sucessfully verified an invalid message")
	}

	// restore the signature and fake a message
	signedMessage.Signature.r.Add(msg.Signature.r, big.NewInt(1))
	signedMessage.Message[0] = 0
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("successfully verified a corrupted message")
	}
}

// test CombinedMessage when the message is nil
