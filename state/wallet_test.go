package state

import (
	"testing"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
)

// TestWalletWeight runs some edge case testing on Wallet.Weight()
func TestWalletCompensationWeight(t *testing.T) {
	var w Wallet

	w.Script = siacrypto.RandomByteSlice(5)
	weight := w.CompensationWeight()
	if weight != 2*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = siacrypto.RandomByteSlice(AtomSize)
	weight = w.CompensationWeight()
	if weight != 2*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = siacrypto.RandomByteSlice(AtomSize + 1)
	weight = w.CompensationWeight()
	if weight != 3*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Sector.Atoms = 12
	weight = w.CompensationWeight()
	if weight != 3*walletAtomMultiplier+uint32(w.Sector.Atoms) {
		t.Error("Wallet weight not properly calculated")
	}

	// Need to rewrite these so that actual updates are inserted into the
	// wallet.
	/*
		w.Sector.UpdateAtoms = 7
		weight = w.CompensationWeight()
		if weight != 3*walletAtomMultiplier+uint32(w.Sector.Atoms)+w.Sector.UpdateAtoms {
			t.Error("Wallet compensation weight not properly calculated.")
		}

		w.Script = nil
		w.Sector.UpdateAtoms = 0
		weight = w.CompensationWeight()
		if weight != walletAtomMultiplier+uint32(w.Sector.Atoms) {
			t.Error("Wallet weight is not being calculated correctly")
		}
	*/
}

// TestInsertLoadSaveRemoveWallet just makes sure that the logic runs without
// error. The components each function called are tested elsewhere in the file.
func TestInsertLoadSaveRemoveWallet(t *testing.T) {
	// Test InsertWallet.
	var s State
	s.SetWalletPrefix(siafiles.TempFilename("TestInsertWallet."))
	var w Wallet
	err := s.InsertWallet(w, true)
	if err != nil {
		t.Error("Trouble while calling InsertWallet", err)
	}

	// Test LoadWallet.
	_, err = s.LoadWallet(w.ID)
	if err != nil {
		t.Error("Trouble while calling LoadWallet", err)
	}

	// Test RemoveWallet, verifying that the wallet is no longer retrievable.
	s.RemoveWallet(w.ID)
	_, err = s.LoadWallet(w.ID)
	if err == nil {
		t.Error("Able to load a removed wallet!")
	}

	// Test SaveWallet, and then make sure that the saved wallet can be loaded.
	w.ID = 25
	err = s.SaveWallet(w)
	if err == nil {
		t.Error("Able to save a wallet that doesn't exist in the wallet tree.")
	}
}
