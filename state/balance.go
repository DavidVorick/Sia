package state

import (
	"siaencoding"
)

// A Balance is a 128-bit integer representing a volume of siacoins.
// The actual 128-bit number will likely be in nano- or femto-siacoins.
type Balance [16]byte

// NewBalance creates a 128-bit Balance from two uint64s.
func NewBalance(upper, lower uint64) (b Balance) {
	uBytes := siaencoding.EncUint64(upper)
	lBytes := siaencoding.EncUint64(lower)
	copy(b[:8], lBytes)
	copy(b[8:], uBytes)
	return
}

// Add performs addition on two Balances.
func (a *Balance) Add(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Add(x, y)))
}

// Subtract performs subtraction on two Balances.
func (a *Balance) Subtract(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Sub(x, y)))
}

// Multiply performs multiplication on two Balances.
func (a *Balance) Multiply(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Mul(x, y)))
}

// Compare returns an integer comparing two Balances.
// It returns 1 if a > b, -1 if a < b, and 0 if a == b
func (a *Balance) Compare(b Balance) int {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	return x.Cmp(y)
}
