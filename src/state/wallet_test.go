package state

import (
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
	weight = w.Weight()
	if weight != 2*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.Script = siacrypto.RandomByteSlice(AtomSize + 1)
	weight = w.Weight()
	if weight != 3*walletAtomMultiplier {
		t.Error("Wallet weight is not being calculated correctly")
	}

	w.SectorSettings.Atoms = 12
	weight = w.Weight()
	if weight != 3*walletAtomMultiplier+uint32(w.SectorSettings.Atoms) {
		t.Error("Wallet weight not properly calculated")
	}

	w.Script = nil
	weight = w.Weight()
	if weight != walletAtomMultiplier+uint32(w.SectorSettings.Atoms) {
		t.Error("Wallet weight is not being calculated correctly")
	}
}

// TestInsertLoadSaveRemoveWallet just makes sure that the logic runs without
// error. The components each function called are tested elsewhere in the file.
func TestInsertLoadSaveRemoveWallet(t *testing.T) {
	// Test InsertWallet.
	var s State
	s.SetWalletPrefix("../../filesCreatedDuringTesting/TestInsertWallet.")
	var w Wallet
	err := s.InsertWallet(w)
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
	s.SaveWallet(w)
	_, err = s.LoadWallet(w.ID)
	if err != nil {
		t.Error("Trouble while calling LoadWallet", err)
	}
}
