package quorum

import (
	"siaencoding"
)

type Balance [16]byte

func NewBalance(upper, lower uint64) (b Balance) {
	uBytes := siaencoding.EncUint64(upper)
	lBytes := siaencoding.EncUint64(lower)
	copy(b[:8], lBytes)
	copy(b[8:], uBytes)
	return
}

// should return an error on overflow
func (a *Balance) Add(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Add(x, y)))
}

// should return an error if b > a
func (a *Balance) Subtract(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Sub(x, y)))
}

// should return an error on overflow
func (a *Balance) Multiply(b Balance) {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	copy(a[:], siaencoding.EncUint128(x.Mul(x, y)))
}

// Compare returns 1 if a > b, -1 if a < b, and 0 if a == b
func (a *Balance) Compare(b Balance) int {
	x := siaencoding.DecUint128(a[:])
	y := siaencoding.DecUint128(b[:])
	return x.Cmp(y)
}
