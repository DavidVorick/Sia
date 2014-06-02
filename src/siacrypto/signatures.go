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
func (pk *PublicKey) Hash() (hash TruncatedHash, err error) {
	if pk == nil {
		err = fmt.Errorf("Cannot hash a nil public key")
		return
	}
	if pk.key.X == nil || pk.key.Y == nil {
		err = fmt.Errorf("Cannot hash an improperly initialized public key")
		return
	}

	combinedKey := append(pk.key.X.Bytes(), pk.key.Y.Bytes()...)
	hash, err = CalculateTruncatedHash(combinedKey)
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

// Return a []byte containing a signature followed by the signed message
func (sm *SignedMessage) CombinedMessage() (combinedMessage []byte, err error) {
	if sm == nil {
		err = fmt.Errorf("Cannot combine a nil signedMessage")
		return
	}

	if sm.Signature.r == nil || sm.Signature.s == nil {
		err = fmt.Errorf("Signature has been improperly initialized")
		return
	}

	combinedMessage = append(sm.Signature.r.Bytes(), sm.Signature.s.Bytes()...)
	combinedMessage = append(combinedMessage, sm.Message...)
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

	r, s, err := ecdsa.Sign(rand.Reader, secKey.key, (message))
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
