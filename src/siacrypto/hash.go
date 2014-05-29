package siacrypto

import (
	"crypto/sha512"
)

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
