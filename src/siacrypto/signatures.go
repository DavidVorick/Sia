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

// takes as input a public key and a signed message
// returns whether the signature is valid or not
func (pk PublicKey) Verify(sig Signature, message []byte) (verified bool) {
	var messagePointer *C.uchar
	messageBytes := make([]byte, len(message)+SignatureSize)
	if len(message) == 0 {
		// can't point to a slice of len 0
		messagePointer = (*C.uchar)(nil)
	} else {
		messagePointer = (*C.uchar)(unsafe.Pointer(&messageBytes[0]))
	}

	var messageLen uint64
	lenPointer := (*C.ulonglong)(unsafe.Pointer(&messageLen))

	signedMessageBytes := append(sig[:], message...)
	signedMessagePointer := (*C.uchar)(unsafe.Pointer(&signedMessageBytes[0]))
	signedMessageLen := C.ulonglong(len(signedMessageBytes))
	pkPointer := (*C.uchar)(unsafe.Pointer(&pk[0]))

	success := C.crypto_sign_open(messagePointer, lenPointer, signedMessagePointer, signedMessageLen, pkPointer)
	verified = success == 0
	return
}

func (pk PublicKey) VerifyObject(sig Signature, obj interface{}) (verified bool, err error) {
	objHash, err := HashObject(obj)
	if err != nil {
		return
	}

	verified = pk.Verify(sig, objHash[:])
	return
}

// Sign takes a secret key and a message, and uses the secret key to sign the
// message,  returning a single SignedMessage struct containing a Message and a
// Signature
func (secKey SecretKey) Sign(message []byte) (sig Signature, err error) {
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

	copy(sig[:], signedMessageBytes)
	return
}

func (secKey SecretKey) SignObject(o interface{}) (s Signature, err error) {
	objectHash, err := HashObject(o)
	if err != nil {
		return
	}

	s, err = secKey.Sign(objectHash[:])
	return
}

// CreateKeyPair needs no input, produces a public key and secret key as output
func CreateKeyPair() (pubKey PublicKey, secKey SecretKey, err error) {
	errorCode := C.crypto_sign_keypair((*C.uchar)(unsafe.Pointer(&pubKey[0])), (*C.uchar)(unsafe.Pointer(&secKey[0])))
	if errorCode != 0 {
		err = fmt.Errorf("Key Creation Failed!")
	}
	return
}
