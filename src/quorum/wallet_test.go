package quorum

import (
	"siacrypto"
	"testing"
)

// TestWalletCoding creates a wallet, fills it with random values, converts it
// to bytes, back to a wallet, and then back to bytes, and then compares each
// against the other to make sure that the process of encoding and decoding
// does not introduce any errors.
func TestWalletCoding(t *testing.T) {
	// Fill out a wallet with completely random values
	w := new(wallet)
	w.upperBalance = siacrypto.RandomUInt64()
	w.lowerBalance = siacrypto.RandomUInt64()
	w.scriptAtoms = uint16(siacrypto.RandomUInt64())
	for i := range w.sectorOverview {
		w.sectorOverview[i].m = siacrypto.RandomByte()
		w.sectorOverview[i].numAtoms = siacrypto.RandomByte()
	}
	randomBytes := siacrypto.RandomByteSlice(scriptPrimerSize)
	copy(w.scriptPrimer[:], randomBytes)

	wBytes := w.bytes()
	wObj := fillWallet(wBytes)
	wConfirm := wObj.bytes()

	if *wBytes != *wConfirm {
		t.Error("wBytes mismatches wConfirm")
	}
	if w.upperBalance != wObj.upperBalance {
		t.Error("Error with upperBalance")
	}
	if w.lowerBalance != wObj.lowerBalance {
		t.Error("Error with lowerBalance")
	}
	if w.scriptAtoms != wObj.scriptAtoms {
		t.Error("Error with scriptAtoms")
	}
	for i := range w.sectorOverview {
		if w.sectorOverview[i].m != wObj.sectorOverview[i].m {
			t.Error("Error with sectorOverview:", i)
		}
		if w.sectorOverview[i].numAtoms != wObj.sectorOverview[i].numAtoms {
			t.Error("Error with sectorOverview:", i)
		}
	}
	if w.scriptPrimer != wObj.scriptPrimer {
		t.Error("Error with scriptPrimer")
	}
}
