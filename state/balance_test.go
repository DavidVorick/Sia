package state

import (
	"testing"
)

func TestOperations(t *testing.T) {
	a := NewBalance(0, ^uint64(0))
	a2 := NewBalance(0, ^uint64(0))
	b := NewBalance(0, 0x00000002)
	c := NewBalance(1, 0x00000001)

	a.Add(b)
	if a.Compare(c) != 0 {
		t.Fatal("addition failed")
	}

	a.Subtract(b)
	if a.Compare(a2) != 0 {
		t.Fatal("subtraction failed")
	}

	a.Add(a)
	a2.Multiply(b)
	if a.Compare(a2) != 0 {
		t.Fatal("multiplication failed")
	}
}
