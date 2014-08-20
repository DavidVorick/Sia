package state

import (
	"testing"
)

// TestOperations tests basic arithmetic operations on Balances.
func TestOperations(t *testing.T) {
	a := NewBalance(^uint64(0))
	a2 := NewStringBalance("18446744073709551615")
	b := NewBalance(2)
	c := Balance{1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}

	if a.Compare(a2) != 0 {
		t.Fatal("NewBalance result does not match NewStringBalance result")
	}

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
