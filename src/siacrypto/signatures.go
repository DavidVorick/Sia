package siacrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

type SignedMessage struct {
	Signature Signature
	Message   []byte
}

// Return a []byte containing both the message and the prepended signature
func (sm *SignedMessage) CombinedMessage() (combinedMessage []byte, err error) {
	if sm == nil {
		err = fmt.Errorf("Cannot combine a nil signedMessage")
		return
	}

	combinedMessage = append(sm.Signature.r.Bytes(), sm.Signature.s.Bytes()...)
	combinedMessage = append(combinedMessage, sm.Message...)

	return
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (publicKey *PublicKey, secretKey SecretKey, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	secretKey.key = *priv
	publicKey.key = priv.PublicKey
	return
}

// Sign takes a secret key and a message, and use the secret key to sign the message.
// Sign returns a single SignedMessage struct containing a Message and a Signature
func (secKey *SecretKey) Sign(message []byte) (signedMessage SignedMessage, err error) {
	if secKey == nil {
		err = fmt.Errorf("Cannot sign using a nil SecretKey")
		return
	}

	if message == nil {
		err = fmt.Errorf("Cannot sign a nil message")
		return
	}

	r, s, err := ecdsa.Sign(rand.Reader, &secKey.key, []byte(message))
	signedMessage.Signature.r = r
	signedMessage.Signature.s = s
	signedMessage.Message = message
	return
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func (pk *PublicKey) Verify(signedMessage *SignedMessage) (verified bool) {
	if pk == nil || signedMessage == nil {
		return false
	}

	verified = ecdsa.Verify(&pk.key, []byte(signedMessage.Message), signedMessage.Signature.r, signedMessage.Signature.s)
	return
}
