package siacrypto

import (
	"crypto/sha512"
)

const (
	// sizes in bytes
	HashSize          int = 64
	TruncatedHashSize int = 32
)

type Hash [HashSize]byte
type TruncatedHash [TruncatedHashSize]byte

// returns the sha512 hash of the input []byte
func CalculateHash(data []byte) (hash Hash, err error) {
	sha := sha512.New()
	sha.Write(data)
	hashSlice := sha.Sum(nil)
	copy(hash[:], hashSlice)
	return
}

// Calls Hash, and then returns only the first TruncatedHashSize bytes
func CalculateTruncatedHash(data []byte) (tHash TruncatedHash, err error) {
	hash, err := CalculateHash(data)
	if err != nil {
		return
	}

	copy(tHash[:], hash[:])
	return
}
