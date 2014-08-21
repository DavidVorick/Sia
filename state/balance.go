package state

import (
	"math/big"

	"github.com/NebulousLabs/Sia/siaencoding"
)

// A Balance is a 128-bit unsigned integer representing a volume of siacoins.
// The actual 128-bit number will likely be in nano- or femto-siacoins.
type Balance [16]byte

// NewBalance creates a Balance from a uint64. Larger Balances can be created
// with the NewStringBalance function, or by modifying the bytes of a Balance
// directly.
func NewBalance(lower uint64) (b Balance) {
	copy(b[:], siaencoding.EncUint64(lower))
	return
}

// NewStringBalance creates a new Balance from a string containing a decimal
// value. For reference, the largest possible decimal value of a Balance is
// 340282366920938463463374607431768211456.
func NewStringBalance(bal string) (b Balance) {
	i := new(big.Int)
	i.SetString(bal, 10)
	copy(b[:], siaencoding.EncUint128(i))
	return
}

// String returns a balance as a string containing a decimal value.
func (a Balance) String() string {
	x := siaencoding.DecUint128(a[:])
	return x.String()
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
