package siacrypto

import (
	"testing"
)

func TestRandomByteSlice(t *testing.T) {
	// run tests with bogus input
	randomByteSlice, err := RandomByteSlice(-3)
	if err == nil {
		t.Error("RandomByteSlice is accepting negative values")
	}

	randomByteSlice, err = RandomByteSlice(400)
	if err != nil {
		t.Fatal(err)
	}

	if len(randomByteSlice) != 400 {
		t.Fatal("Incorrect number of bytes generated!")
	}

	randomByteSlice, err = RandomByteSlice(0)
	if len(randomByteSlice) != 0 {
		t.Error("unspecified behavoir when calling RandomByteSlice(0)")
	}

	// add a statistical test to verify that the data appears random
}

// Benchmark function for RandomByteSlice

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
