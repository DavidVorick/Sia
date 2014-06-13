package quorum

import (
	"bytes"
	"siacrypto"
	"testing"
)

// TestWalletCoding creates a wallet, fills it with random values, converts it
// to bytes, back to a wallet, and then back to bytes, and then compares each
// against the other to make sure that the process of encoding and decoding
// does not introduce any errors.
func TestWalletCoding(t *testing.T) {
	// Fill out a wallet with completely random values
	w := new(Wallet)
	w.Balance = NewBalance(siacrypto.RandomUInt64(), siacrypto.RandomUInt64())
	for i := range w.sectorOverview {
		w.sectorOverview[i].m = siacrypto.RandomByte()
		w.sectorOverview[i].atoms = siacrypto.RandomByte()
	}
	randomBytes := siacrypto.RandomByteSlice(400)
	copy(w.script, randomBytes)

	wBytes, err := w.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	var wObj Wallet
	err = wObj.GobDecode(wBytes)
	if err != nil {
		t.Fatal(err)
	}
	wConfirm, err := wObj.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(wBytes, wConfirm) != 0 {
		t.Error("wBytes mismatches wConfirm")
	}
	if w.Balance != wObj.Balance {
		t.Error("Error with upperBalance")
	}
	for i := range w.sectorOverview {
		if w.sectorOverview[i].m != wObj.sectorOverview[i].m {
			t.Error("Error with sectorOverview:", i)
		}
		if w.sectorOverview[i].atoms != wObj.sectorOverview[i].atoms {
			t.Error("Error with sectorOverview:", i)
		}
	}
	if bytes.Compare(w.script, wObj.script) != 0 {
		t.Error("Script mismatch")
	}
}
