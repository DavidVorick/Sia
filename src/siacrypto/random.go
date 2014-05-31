package siacrypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// RandomByteSlice takes an int as input and returns a []byte of length int that
// is full of random values
func RandomByteSlice(numBytes int) (randomBytes []byte, err error) {
	if numBytes < 0 {
		err = fmt.Errorf("RandomByteSlice: cannot generate a negative number of bytes")
		return
	}

	randomBytes = make([]byte, numBytes)
	rand.Read(randomBytes)
	return
}

// RandomInt() generates a random int [0, ceiling)
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
