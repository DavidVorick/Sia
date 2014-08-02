package siacrypto

// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
	"errors"
)

const (
	PublicKeySize = 32
	SecretKeySize = 64
	SignatureSize = 64
)

type PublicKey [PublicKeySize]byte
type SecretKey [SecretKeySize]byte
type Signature [SignatureSize]byte

// Verify returns whether a message was signed by the public key 'pk'.
func (pk PublicKey) Verify(sig Signature, message []byte) bool {
	messageBytes := make([]byte, len(message)+SignatureSize)
	messagePointer := (*C.uchar)(&messageBytes[0])

	var messageLen uint64
	lenPointer := (*C.ulonglong)(&messageLen)

	signedMessageBytes := append(sig[:], message...)
	signedMessagePointer := (*C.uchar)(&signedMessageBytes[0])
	signedMessageLen := C.ulonglong(len(signedMessageBytes))
	pkPointer := (*C.uchar)(&pk[0])

	errorCode := C.crypto_sign_open(messagePointer, lenPointer, signedMessagePointer, signedMessageLen, pkPointer)
	return errorCode == 0
}

// VerifyObject returns whether an object was signed by the public key 'pk'.
// It does so by first marshalling the object, and then passing the result to Verify().
func (pk PublicKey) VerifyObject(sig Signature, obj interface{}) (verified bool, err error) {
	objHash, err := HashObject(obj)
	if err != nil {
		return
	}

	verified = pk.Verify(sig, objHash[:])
	return
}

// Sign returns the signature of a byte slice.
func (sk SecretKey) Sign(message []byte) (sig Signature, err error) {
	if message == nil {
		err = errors.New("cannot sign a nil message")
		return
	}

	signedMessageBytes := make([]byte, len(message)+SignatureSize)
	signedMessagePointer := (*C.uchar)(&signedMessageBytes[0])

	var signatureLen uint64
	lenPointer := (*C.ulonglong)(&signatureLen)

	var messagePointer *C.uchar
	if len(message) == 0 {
		// can't point to a slice of len 0
		messagePointer = (*C.uchar)(nil)
	} else {
		messageBytes := []byte(message)
		messagePointer = (*C.uchar)(&messageBytes[0])
	}

	messageLen := C.ulonglong(len(message))
	skPointer := (*C.uchar)(&sk[0])

	signErr := C.crypto_sign(signedMessagePointer, lenPointer, messagePointer, messageLen, skPointer)
	if signErr != 0 {
		err = errors.New("call to crypto_sign failed")
		return
	}

	copy(sig[:], signedMessageBytes)
	return
}

// SignObject returns the signature of an object's hash.
func (sk SecretKey) SignObject(obj interface{}) (sig Signature, err error) {
	objectHash, err := HashObject(obj)
	if err != nil {
		return
	}

	return sk.Sign(objectHash[:])
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (pubKey PublicKey, secKey SecretKey, err error) {
	errorCode := C.crypto_sign_keypair((*C.uchar)(&pubKey[0]), (*C.uchar)(&secKey[0]))
	if errorCode != 0 {
		err = errors.New("call to crypto_sign_keypair failed")
	}
	return
}
