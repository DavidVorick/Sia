package siacrypto

// Eventually, hash.go needs to be moved so that it uses libsodium

import (
	"crypto/sha512"
)

const (
	HashSize int = 32 // in bytes
)

type Hash [HashSize]byte

// returns the first 256 bytes of the sha512 hash of the input []byte
func CalculateHash(data []byte) (hash Hash) {
	sha := sha512.New()
	sha.Write(data)
	hashSlice := sha.Sum(nil)
	copy(hash[:], hashSlice)
	return
}
