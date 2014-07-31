package siacrypto

// Should be moved to using libsodium

import (
	"crypto/rand"
	"fmt"
	"siaencoding"
)

// RandomByte returns a random byte
func RandomByte() byte {
	randomByte := make([]byte, 1)
	rand.Read(randomByte)
	return randomByte[0]
}

// RandomByteSlice returns a slice of random bytes
func RandomByteSlice(numBytes int) (randomBytes []byte) {
	if numBytes < 0 {
		randomBytes = make([]byte, 0)
		return
	}

	randomBytes = make([]byte, numBytes)
	rand.Read(randomBytes)
	return
}

// RandomInt generates a random int [0, ceiling)
func RandomInt(ceiling int) (randInt int, err error) {
	if ceiling < 1 {
		err = fmt.Errorf("RandomInt: input must be greater than 0")
		return
	}

	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)

	randInt = int(siaencoding.DecUint32(randomBytes)) % ceiling
	return
}

// RandomUint16 returns a random uint16.
// It accomplishes this by feeding 2 bytes of random data to a binary decoder.
func RandomUint16() uint16 {
	randomBytes := make([]byte, 2)
	rand.Read(randomBytes)
	return siaencoding.DecUint16(randomBytes)
}

// RandomUint64 returns a random uint64.
// It uses the same process as RandomUint16.
func RandomUint64() (randInt uint64) {
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return siaencoding.DecUint64(randomBytes)
}
