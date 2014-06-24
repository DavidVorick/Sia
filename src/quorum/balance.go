package quorum

import (
	"fmt"
	"siaencoding"
)

type Balance struct {
	upperBalance uint64
	lowerBalance uint64
}

func NewBalance(upper, lower uint64) Balance {
	return Balance{upper, lower}
}

func (a *Balance) Add(b Balance) {
	a.upperBalance += b.upperBalance
	if ^uint64(0)-a.lowerBalance >= b.lowerBalance {
		a.lowerBalance += b.lowerBalance
	} else {
		a.upperBalance += 1
		if a.lowerBalance < b.lowerBalance {
			a.lowerBalance = a.lowerBalance - (^uint64(0) - b.lowerBalance)
		} else {
			a.lowerBalance = b.lowerBalance - (^uint64(0) - a.lowerBalance)
		}
	}
}

func (a *Balance) Subtract(b Balance) {
	a.upperBalance -= b.upperBalance
	if a.lowerBalance < b.lowerBalance {
		a.upperBalance -= 1
		a.lowerBalance = ^uint64(0) - (b.lowerBalance - a.lowerBalance)
	} else {
		a.lowerBalance -= b.lowerBalance
	}
}

// Luke does stuff here
func (a *Balance) Multiply(i uint32) {
}

// Compare returns true if a is greater than or equal to b
func (a *Balance) Compare(b Balance) bool {
	if a.upperBalance < b.upperBalance {
		return false
	}
	if a.upperBalance == b.upperBalance && a.lowerBalance < b.lowerBalance {
		return false
	}

	return true
}

func (b *Balance) GobEncode() ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("Cannot encode nil Balance")
	}

	upperBytes := siaencoding.EncUint64(b.upperBalance)
	lowerBytes := siaencoding.EncUint64(b.lowerBalance)
	return append(upperBytes, lowerBytes...), nil
}

func (b *Balance) GobDecode(bytes []byte) error {
	if b == nil {
		return fmt.Errorf("Cannot decode into nil Balance")
	}
	if len(bytes) != 16 {
		return fmt.Errorf("Invalid encoded Balance!")
	}

	b.upperBalance = siaencoding.DecUint64(bytes[:8])
	b.lowerBalance = siaencoding.DecUint64(bytes[8:])
	return nil
}
