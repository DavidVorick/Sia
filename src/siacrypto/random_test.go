package siacrypto

import (
	"testing"
)

func TestRandomByteSlice(t *testing.T) {
	// run tests with bogus input
	randomByteSlice := RandomByteSlice(-3)
	if len(randomByteSlice) != 0 {
		t.Error("RandomByteSlice input with a negative number should return a len 0 byte slice")
	}
	randomByteSlice = RandomByteSlice(400)
	if len(randomByteSlice) != 400 {
		t.Error("Incorrect number of bytes generated!")
	}
	randomByteSlice = RandomByteSlice(0)
	if len(randomByteSlice) != 0 {
		t.Error("unspecified behavoir when calling RandomByteSlice(0)")
	}
}

func TestRandomInt(t *testing.T) {
	// test 1 as a ceiling in range [0, 1)
	zero, err := RandomInt(1)
	if err != nil {
		t.Fatal(err)
	}
	if zero != 0 {
		t.Fatal("Expecting rng to produce 0!")
	}

	zero, err = RandomInt(0)
	if err == nil {
		t.Error("Expecting RandomInt(0) to produce an error!")
	}

	// a series of tests that stastically checks for randomness
}

// Benchmark function for RandomInt
