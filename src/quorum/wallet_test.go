package quorum

import (
	"siacrypto"
	"testing"
)

func TestIdEncoding(t *testing.T) {
	randomBytes, err := siacrypto.RandomByteSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	var w0 walletHandle
	copy(w0[:], randomBytes)
	w0ID := w0.id()
	w0Handle := w0ID.handle()
	w0Confirm := w0Handle.id()

	if w0 != w0Handle {
		t.Error("Encoding Mismatch:", w0, ":", w0Handle)
	}
	if w0ID != w0Confirm {
		t.Error("Encoding Mismatch:", w0ID, ":", w0Confirm)
	}
}

func TestWalletCoding(t *testing.T) {
	// Fill out a wallet with completely random values
	w := new(wallet)
	max := int64(^uint64(0) >> 1)
	maxu16 := int64(^uint16(0))
	w.upperBalance = siacrypto.RandomInt64(max)
	w.lowerBalance = siacrypto.RandomInt64(max)
	w.scriptAtoms = uint16(siacrypto.RandomInt64(maxu16))
	for i := range w.sectorOverview {
		w.sectorOverview[i].m = siacrypto.RandomByte()
		w.sectorOverview[i].numAtoms = siacrypto.RandomByte()
	}
	randomBytes, _ := siacrypto.RandomByteSlice(scriptPrimerSize)
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
