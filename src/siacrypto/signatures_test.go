package siacrypto

import (
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
}

func TestSecretKeyCompare(t *testing.T) {
	// compare nil public keys
	var sk0 *SecretKey
	var sk1 *SecretKey
	compare := sk0.Compare(sk1)
	if compare {
		t.Error("Comparing nil public keys return true")
	}

	// compare when one public key is nil
	_, sk0, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = sk0.Compare(sk1)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}
	compare = sk1.Compare(sk0)
	if compare {
		t.Error("Comparing a nil public key returns true")
	}

	// compare unequal public keys
	_, sk1, err = CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	compare = sk0.Compare(sk1)
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
}

func TestSecretKeyEncoding(t *testing.T) {
	// Encode and Decode nil values
	var sk *SecretKey
	_, _ = sk.GobEncode() // checking for panics
	sk = new(SecretKey)
	_, _ = sk.GobEncode() // checking for panics

	_ = sk.GobDecode(nil) // checking for panics

	// Encode and then Decode, see if the results are identical
	_, secKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	eSecKey, err := secKey.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = sk.GobDecode(eSecKey)
	if err != nil {
		t.Fatal(err)
	}
	compare := sk.Compare(secKey)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
	}
	compare = secKey.Compare(sk)
	if !compare {
		t.Error("Encoded and then decoded key not equal")
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
	verified := publicKey.Verify(&msg)
	if !verified {
		t.Error("Signed empty message did not verify!")
	}

	if testing.Short() {
		t.Skip()
	}

	// verify empty message when signature is bad
	msg.Signature[0] = ^msg.Signature[0] // flip the bits on the first byte to guarantee corruption
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

	// sign the message
	signedMessage, err := secretKey.Sign(RandomByteSlice(20))
	if err != nil {
		return
	}

	// verify the signature
	verification := publicKey.Verify(&signedMessage)
	if !verification {
		t.Error("failed to verify a valid message")
	}

	// verify an imposter signature
	signedMessage.Signature[0] = ^signedMessage.Signature[0]
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("sucessfully verified an invalid message")
	}

	// restore the signature and fake a message
	signedMessage.Signature[0] = ^signedMessage.Signature[0]
	signedMessage.Message[0] = ^signedMessage.Message[0]
	verification = publicKey.Verify(&signedMessage)
	if verification {
		t.Error("successfully verified a corrupted message")
	}
}

func TestMessageEncoding(t *testing.T) {
	publicKey, secretKey, err := CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("test")
	signedMessage, err := secretKey.Sign(message)
	if err != nil {
		t.Fatal(err)
	}
	verified := publicKey.Verify(&signedMessage)
	if !verified {
		t.Error("verification failed")
	}

	gobSm, err := signedMessage.GobEncode()
	if err != nil {
		t.Error(err)
	}

	var decSm SignedMessage
	err = decSm.GobDecode(gobSm)
	if err != nil {
		t.Fatal(err)
	}
	verified = publicKey.Verify(&decSm)
	if !verified {
		t.Error("signed message corrupted during encode/decode")
	}
}
