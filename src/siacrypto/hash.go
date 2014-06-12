package siacrypto

import (
	"crypto/sha512"
)

const (
	HashSize int = 32 // in bytes
)

type Hash [HashSize]byte

// returns the sha512 hash of the input []byte
func CalculateHash(data []byte) (hash Hash, err error) {
	sha := sha512.New()
	sha.Write(data)
	hashSlice := sha.Sum(nil)
	copy(hash[:], hashSlice)
	return
}
