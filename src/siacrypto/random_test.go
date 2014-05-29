package siacrypto

import (
	"testing"
)

func TestRandomByteSlice(t *testing.T) {
	randomByteSlice, err := RandomByteSlice(400)
	if err != nil {
		t.Fatal(err)
	}

	if len(randomByteSlice) != 400 {
		t.Fatal("Incorrect number of bytes generated!")
	}

	// add a statistical test to verify that the data appears
	// random

	// there should be a longer test, and perhaps a benchmark
	// which generates a very large random slice
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

	// a series of tests that stastically checks for randomness
}
