package siacrypto

// Eventually, hash.go needs to be moved so that it uses libsodium

import (
	"crypto/sha512"
	"siaencoding"
)

const (
	HashSize int = 32 // in bytes
)

type Hash [HashSize]byte

// HashBytes returns the first 32 bytes of the sha512 hash of the input.
func HashBytes(data []byte) (h Hash) {
	hash512 := sha512.Sum512(data)
	copy(h[:], hash512[:])
	return
}

// HashObject converts an object to a byte slice and returns the hash of the byte slice.
func HashObject(obj interface{}) (h Hash, err error) {
	bytes, err := siaencoding.Marshal(obj)
	if err != nil {
		return
	}

	h = HashBytes(bytes)
	return
}
