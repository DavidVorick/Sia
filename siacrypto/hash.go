package siacrypto

// Eventually, hash.go needs to be moved so that it uses libsodium

import (
	"crypto/sha512"

	"github.com/NebulousLabs/Sia/siaencoding"
)

const (
	// HashSize is the size of a Hash in bytes
	HashSize int = 32
)

// A Hash is the first 32 bytes of a sha512 checksum.
type Hash [HashSize]byte

// HashBytes returns the first 32 bytes of the sha512 checksum of the input.
func HashBytes(data []byte) (h Hash) {
	hash512 := sha512.Sum512(data)
	copy(h[:], hash512[:])
	return
}

// HashObject converts an object to a byte slice and returns the hash of the
// byte slice.
func HashObject(obj interface{}) (h Hash, err error) {
	bytes, err := siaencoding.Marshal(obj)
	if err != nil {
		return
	}

	h = HashBytes(bytes)
	return
}
