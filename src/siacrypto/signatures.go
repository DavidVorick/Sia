package siacrypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"
)

// The underlying variables to the keys & signatures should not be exported
type PublicKey struct {
	key *ecdsa.PublicKey
}
type SecretKey struct {
	key *ecdsa.PrivateKey
}
type Signature struct {
	r, s *big.Int
}

// A SignedMessage contains a message and a signature of the message
type SignedMessage struct {
	Signature Signature
	Message   []byte
}

// Creates a deterministic hash of a public key
func (pk *PublicKey) Hash() (hash Hash, err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot hash a nil public key")
		return
	}
	if pk.key.X == nil || pk.key.Y == nil {
		err = fmt.Errorf("Cannot hash an improperly initialized public key")
		return
	}

	combinedKey := append(pk.key.X.Bytes(), pk.key.Y.Bytes()...)
	hash = CalculateHash(combinedKey)
	return
}

// Compare returns true if the keys are composed of the same integer values
// Compare returns false if any sub-value is nil
func (pk0 *PublicKey) Compare(pk1 *PublicKey) bool {
	// check for nil values
	if pk0 == nil || pk1 == nil {
		return false
	}
	if pk0.key == nil || pk1.key == nil {
		return false
	}
	if pk0.key.X == nil || pk0.key.Y == nil || pk1.key.X == nil || pk1.key.Y == nil {
		return false
	}

	cmp := pk0.key.X.Cmp(pk1.key.X)
	if cmp != 0 {
		return false
	}

	cmp = pk0.key.Y.Cmp(pk1.key.Y)
	if cmp != 0 {
		return false
	}

	return true
}

// Compare returns true if the keys are composed of the same integer values
// Compare returns false if any sub-value is nil
func (sk0 *SecretKey) Compare(sk1 *SecretKey) bool {
	// check for nil values
	if sk0 == nil || sk1 == nil {
		return false
	}
	if sk0.key == nil || sk1.key == nil {
		return false
	}
	if sk0.key.X == nil || sk0.key.Y == nil || sk1.key.X == nil || sk1.key.Y == nil {
		return false
	}

	cmp := sk0.key.X.Cmp(sk1.key.X)
	if cmp != 0 {
		return false
	}

	cmp = sk0.key.Y.Cmp(sk1.key.Y)
	if cmp != 0 {
		return false
	}

	return true
}

func (pk *PublicKey) GobEncode() (gobPk []byte, err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}
	if pk.key == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}
	if pk.key.X == nil || pk.key.Y == nil {
		err = fmt.Errorf("public key not properly initialized")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(pk.key.X)
	if err != nil {
		return
	}
	err = encoder.Encode(pk.key.Y)
	if err != nil {
		return
	}
	gobPk = w.Bytes()
	return
}

func (pk *PublicKey) GobDecode(gobPk []byte) (err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot decode into nil value")
		return
	}

	pk.key = new(ecdsa.PublicKey)

	r := bytes.NewBuffer(gobPk)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&pk.key.X)
	if err != nil {
		return
	}
	err = decoder.Decode(&pk.key.Y)
	if err != nil {
		return
	}
	pk.key.Curve = elliptic.P521() // might there be a way to make this const?
	return
}

func (sk *SecretKey) GobEncode() (gobSk []byte, err error) {
	if sk == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}
	if sk.key == nil {
		err = fmt.Errorf("Cannot encode a nil value")
		return
	}
	if sk.key.X == nil || sk.key.Y == nil {
		err = fmt.Errorf("secret key not properly initialized")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(sk.key.X)
	if err != nil {
		return
	}
	err = encoder.Encode(sk.key.Y)
	if err != nil {
		return
	}
	gobSk = w.Bytes()
	return
}

func (sk *SecretKey) GobDecode(gobSk []byte) (err error) {
	if sk == nil {
		err = fmt.Errorf("Cannot decode into nil value")
		return
	}

	sk.key = new(ecdsa.PrivateKey)

	r := bytes.NewBuffer(gobSk)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&sk.key.X)
	if err != nil {
		return
	}
	err = decoder.Decode(&sk.key.Y)
	if err != nil {
		return
	}
	sk.key.Curve = elliptic.P521() // might there be a way to make this const?
	return
}

func (sm *SignedMessage) GobEncode() (gobSm []byte, err error) {
	if sm == nil {
		err = fmt.Errorf("Cannot encode a nil SignedMessage")
		return
	}

	if sm.Signature.r == nil || sm.Signature.s == nil {
		err = fmt.Errorf("Signature not properly initialized")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(sm.Signature.r.Bytes())
	if err != nil {
		return
	}
	err = encoder.Encode(sm.Signature.s.Bytes())
	if err != nil {
		return
	}
	err = encoder.Encode(sm.Message)
	if err != nil {
		return
	}

	gobSm = w.Bytes()
	return
}

func (sm *SignedMessage) GobDecode(gobSm []byte) (err error) {
	if sm == nil {
		err = fmt.Errorf("Cannot decode into a nil SignedMessage")
		return
	}
	if sm.Signature.r == nil && sm.Signature.s == nil {
		sm.Signature.r = new(big.Int)
		sm.Signature.s = new(big.Int)
	}

	r := bytes.NewBuffer(gobSm)
	decoder := gob.NewDecoder(r)
	var rBytes []byte
	err = decoder.Decode(&rBytes)
	if err != nil {
		return
	}
	sm.Signature.r.SetBytes(rBytes)

	var sBytes []byte
	err = decoder.Decode(&sBytes)
	if err != nil {
		return
	}
	sm.Signature.s.SetBytes(sBytes)

	err = decoder.Decode(&sm.Message)
	if err != nil {
		return
	}

	return
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (publicKey *PublicKey, secretKey *SecretKey, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return
	}

	secretKey = new(SecretKey)
	secretKey.key = priv
	publicKey = new(PublicKey)
	publicKey.key = &priv.PublicKey
	return
}

// Sign takes a secret key and a message, and use the secret key to sign the message.
// Sign returns a single SignedMessage struct containing a Message and a Signature
func (secKey *SecretKey) Sign(message []byte) (signedMessage SignedMessage, err error) {
	if secKey == nil {
		err = fmt.Errorf("Cannot sign using a nil SecretKey")
		return
	}
	if secKey.key == nil {
		err = fmt.Errorf("Secret Key not properly initialized")
		return
	}
	if secKey.key.X == nil || secKey.key.Y == nil {
		err = fmt.Errorf("Secret Key not properly initialized")
		return
	}
	if message == nil {
		err = fmt.Errorf("Cannot sign a nil message")
		return
	}

	r, s, err := ecdsa.Sign(rand.Reader, secKey.key, message)
	signedMessage.Signature.r = r
	signedMessage.Signature.s = s
	signedMessage.Message = message
	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func (pk *PublicKey) Verify(signedMessage *SignedMessage) (verified bool) {
	if pk == nil || signedMessage == nil {
		return
	}
	if pk.key == nil {
		return
	}
	if pk.key.X == nil || pk.key.Y == nil {
		return
	}

	verified = ecdsa.Verify(pk.key, []byte(signedMessage.Message), signedMessage.Signature.r, signedMessage.Signature.s)
	return
}

func (s *Signature) GobEncode() (gobSig []byte, err error) {
	if s.r == nil || s.s == nil {
		err = fmt.Errorf("Cannot encode nil signature - sigature improperly initialized")
		return
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(s.r)
	if err != nil {
		return
	}
	err = encoder.Encode(s.s)
	if err != nil {
		return
	}
	gobSig = w.Bytes()
	return
}

func (s *Signature) GobDecode(gobSig []byte) (err error) {
	if s == nil {
		err = fmt.Errorf("Cannot decode into a nil value")
		return
	}

	r := bytes.NewBuffer(gobSig)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&s.r)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.s)
	return
}
