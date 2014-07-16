package siacrypto

// Should be moved to using libsodium

import (
	"crypto/rand"
	"fmt"
	"math/big"
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

	bigInt := big.NewInt(int64(ceiling))
	randBig, err := rand.Int(rand.Reader, bigInt)
	if err != nil {
		return
	}
	randInt = int(randBig.Int64())
	return
}

// RandomUint16 returns a random uint16, no ceiling
func RandomUint16() (randInt uint16) {
	maxint64 := int64(^uint64(0) >> 1)
	bigInt := big.NewInt(maxint64)
	randBig, err := rand.Int(rand.Reader, bigInt)
	if err != nil {
		return
	}
	randInt = uint16(randBig.Int64())
	return
}

// RandomUint64() generates a random uint64 of any value
func RandomUint64() (randInt uint64) {
	maxint64 := int64(^uint64(0) >> 1)
	bigInt := big.NewInt(maxint64)
	randBig, err := rand.Int(rand.Reader, bigInt)
	if err != nil {
		return
	}
	randInt = uint64(randBig.Int64())
	randBig, err = rand.Int(rand.Reader, bigInt)
	if err != nil {
		return
	}
	randInt += uint64(randBig.Int64())
	return
}
