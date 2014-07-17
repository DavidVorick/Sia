package siacrypto

// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	PublicKeySize = 32
	SecretKeySize = 64
	SignatureSize = 64
)

// The underlying variables to the keys & signatures should not be exported
type PublicKey [PublicKeySize]byte
type SecretKey [SecretKeySize]byte
type Signature [SignatureSize]byte

// The Keypair contains a PublicKey and its corresponding SecretKey
type Keypair struct {
	PK *PublicKey
	SK *SecretKey
}

// A SignedMessage contains a message and a signature of the message
type SignedMessage struct {
	Signature Signature
	Message   []byte
}

// Creates a deterministic hash of a public key
func (pk PublicKey) Hash() (hash Hash) {
	hash = CalculateHash(pk[:])
	return
}

// Creates a deterministic hash of a secret key
func (sk SecretKey) Hash() (hash Hash) {
	hash = CalculateHash(sk[:])
	return
}

// Compare returns true only if the public keys are non-nil and equivalent
func (pk0 PublicKey) Compare(pk1 PublicKey) bool {
	return pk0 == pk1
}

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func (pk PublicKey) Verify(signedMessage SignedMessage) (verified bool) {
	var messagePointer *C.uchar
	messageBytes := make([]byte, len(signedMessage.Message)+SignatureSize)
	if len(signedMessage.Message) == 0 {
		// can't point to a slice of len 0
		messagePointer = (*C.uchar)(nil)
	} else {
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes[0]))
	}

	var messageLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&messageLen))

	signedMessageBytes := signedMessage.CombinedMessage()
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))
	signedMessageLen := C.ulonglong(len(signedMessageBytes))
	pkPointer := (*C.uchar)(unsafe.Pointer(&pk[0]))

	success := C.crypto_sign_open(messagePointer, lenPointer, signedMessagePointer, signedMessageLen, pkPointer)
	verified = success == 0
	return
}

// Compare returns true if the keys are composed of the same integer values
// Compare returns false if any sub-value is nil
func (sk0 SecretKey) Compare(sk1 SecretKey) bool {
	return sk0 == sk1
}

// Sign takes a secret key and a message, and uses the secret key to sign the
// message,  returning a single SignedMessage struct containing a Message and a
// Signature
func (secKey SecretKey) Sign(message []byte) (signedMessage SignedMessage, err error) {
	if message == nil {
		err = fmt.Errorf("Cannot sign a nil message")
		return
	}

	signedMessageBytes := make([]byte, len(message)+SignatureSize)
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))

	var signatureLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&signatureLen))

	var messagePointer *C.uchar
	if len(message) == 0 {
		// can't point to a slice of len 0
		messagePointer = (*C.uchar)(nil)
	} else {
		messageBytes := []byte(message)
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes[0]))
	}

	messageLen := C.ulonglong(len(message))
	sigPointer := (*C.uchar)(unsafe.Pointer(&secKey[0]))

	signErr := C.crypto_sign(signedMessagePointer, lenPointer, messagePointer, messageLen, sigPointer)
	if signErr != 0 {
		err = fmt.Errorf("Signature Failed!")
		return
	}

	signedMessage.Message = message
	copy(signedMessage.Signature[:], signedMessageBytes[:SignatureSize])
	return
}

func (sm SignedMessage) CombinedMessage() []byte {
	return append(sm.Signature[:], sm.Message...)
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (pubKey PublicKey, secKey SecretKey, err error) {
	errorCode := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&pubKey[0])), (*C.uchar)(unsafe.Pointer(&secKey[0])))
	if errorCode != 0 {
		err = fmt.Errorf("Key Creation Failed!")
	}
	return
}
