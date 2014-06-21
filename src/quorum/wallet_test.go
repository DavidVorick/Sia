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
	w := &Wallet{
		id:          WalletID(siacrypto.RandomUInt64()),
		Balance:     NewBalance(siacrypto.RandomUInt64(), siacrypto.RandomUInt64()),
		sectorAtoms: siacrypto.RandomUInt16(),
		sectorM:     35,
		scriptAtoms: siacrypto.RandomUInt16(),
		script:      siacrypto.RandomByteSlice(45),
	}
	copy(w.walletHash[:], siacrypto.RandomByteSlice(siacrypto.HashSize))
	copy(w.sectorHash[:], siacrypto.RandomByteSlice(siacrypto.HashSize))

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
		t.Error("Error with balance")
	}
	if w.sectorAtoms != wObj.sectorAtoms {
		t.Error("Error with sectorAtoms")
	}
	if w.sectorM != wObj.sectorM {
		t.Error("Error with sectorM")
	}
	if w.sectorHash != wObj.sectorHash {
		t.Error("Error with sectorHash")
	}
	if w.scriptAtoms != wObj.scriptAtoms {
		t.Error("Error with scriptAtoms")
	}
	if bytes.Compare(w.script, wObj.script) != 0 {
		t.Error("Script mismatch")
	}
}
