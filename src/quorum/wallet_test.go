package quorum

import (
	"bytes"
	"siacrypto"
	"testing"
)

// TestWalletWeight runs some edge case testing on Wallet.Weight()
func TestWalletWeight(t *testing.T) {
	var w Wallet

	w.Script = siacrypto.RandomByteSlice(5)
	weight := w.Weight()
	if weight != 2*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = siacrypto.RandomByteSlice(AtomSize)
	weight := w.Weight()
	if weight != 2*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = siacrypto.RandomByteSlice(AtomSize + 1)
	weight := w.Weight()
	if weight != 3*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = nil
	weight := w.Weight()
	if weight != walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}
}

func TestWalletCoding(t *testing.T) {
}
