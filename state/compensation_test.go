package state

import (
	"siacrypto"
	"siafiles"
	"testing"
)

// TestExecuteCompensation creates 3 wallets and a sibling, and tests
// compensation for 'StoragePrice' = 1. After verifying that each wallet was
// charged the correct amount and verifying that each sibling received the
// correct amount, 'StoragePrice' is set to 3 and a second sibling is added.
// ExecuteCompensation() is run again, and the balances are verified again.
// Finally, ExecuteCompensation() is run a third time, which knocks the third
// wallet down to a 0 balance. TestExecuteCompensation then verifies that the
// wallet has been removed from the quorum.
func TestExecuteCompensation(t *testing.T) {
	// Initialize the state and set the storage price to 1.
	var s State
	s.SetWalletPrefix(siafiles.TempFilename("TestExecuteCompensation."))
	s.Metadata.StoragePrice = NewBalance(0, 1)

	// Create 3 wallets, a base wallet, a wallet with a script, and a wallet with
	// a sector.
	w0 := Wallet{
		ID:      0,
		Balance: NewBalance(0, 100),
	}
	w1 := Wallet{
		ID:      1,
		Balance: NewBalance(0, 100),
		Script:  siacrypto.RandomByteSlice(8),
	}
	w2 := Wallet{
		ID:      2,
		Balance: NewBalance(0, 100),
		SectorSettings: SectorSettings{
			Atoms: 10,
		},
	}

	// Insert the wallets into the State.
	s.InsertWallet(w0)
	s.InsertWallet(w1)
	s.InsertWallet(w2)

	// Add a sibling to the state with its own wallet.
	sib0Wallet := Wallet{
		ID:      3,
		Balance: NewBalance(0, 100),
	}
	s.InsertWallet(sib0Wallet)
	sib0 := Sibling{
		Active:   true,
		Index:    0,
		WalletID: 3,
	}
	s.Metadata.Siblings[sib0.Index] = sib0

	// Run 'ExecuteCompensation' and see that all the wallets were properly
	// deducted.
	s.ExecuteCompensation()

	w0, err := s.LoadWallet(0)
	if err != nil {
		t.Fatal(err)
	}
	w0ExpectedBalance := NewBalance(0, 100-walletAtomMultiplier)
	if w0.Balance.Compare(w0ExpectedBalance) != 0 {
		t.Error("w0 did not have the correct balance after compensation", w0.Balance)
	}

	w1, err = s.LoadWallet(1)
	if err != nil {
		t.Fatal(err)
	}
	w1ExpectedBalance := NewBalance(0, 100-walletAtomMultiplier*2)
	if w1.Balance.Compare(w1ExpectedBalance) != 0 {
		t.Error("w1 did not have the expected balance after compensation", w1.Balance)
	}

	w2, err = s.LoadWallet(2)
	if err != nil {
		t.Fatal(err)
	}
	w2ExpectedBalance := NewBalance(0, 100-walletAtomMultiplier-10)
	if w2.Balance.Compare(w2ExpectedBalance) != 0 {
		t.Error("w1 did not have the expected balance after compensation")
	}

	// Check that the sibling was properly compensated.
	sib0Wallet, err = s.LoadWallet(3)
	if err != nil {
		t.Fatal(err)
	}
	sib0ExpectedBalance := NewBalance(0, 100+walletAtomMultiplier*4+10)
	if sib0Wallet.Balance.Compare(sib0ExpectedBalance) != 0 {
		t.Error("sibling did not have expected balance after compensation")
	}

	// Add a second sibling, and increase the storage price.
	sib1Wallet := Wallet{
		ID:      4,
		Balance: NewBalance(0, 100),
	}
	s.InsertWallet(sib1Wallet)
	sib1 := Sibling{
		Active:   true,
		Index:    1,
		WalletID: 4,
	}
	s.Metadata.Siblings[sib1.Index] = sib1
	s.Metadata.StoragePrice = NewBalance(0, 3)

	// Run 'ExecuteCompensation' and see that all wallets were properly deducted.
	s.ExecuteCompensation()

	w0, err = s.LoadWallet(0)
	if err != nil {
		t.Fatal(err)
	}
	w0ExpectedBalance = NewBalance(0, 100-7*walletAtomMultiplier)
	if w0.Balance.Compare(w0ExpectedBalance) != 0 {
		t.Error("w0 did not have the correct balance after compensation")
	}

	w1, err = s.LoadWallet(1)
	if err != nil {
		t.Fatal(err)
	}
	w1ExpectedBalance = NewBalance(0, 100-7*walletAtomMultiplier*2)
	if w1.Balance.Compare(w1ExpectedBalance) != 0 {
		t.Error("w1 did not have the expected balance after compensation")
	}

	w2, err = s.LoadWallet(2)
	if err != nil {
		t.Fatal(err)
	}
	w2ExpectedBalance = NewBalance(0, 100-7*(walletAtomMultiplier+10))
	if w2.Balance.Compare(w2ExpectedBalance) != 0 {
		t.Error("w1 did not have the expected balance after compensation")
	}

	// Check that the siblings were properly compensated.
	sib0Wallet, err = s.LoadWallet(3)
	if err != nil {
		t.Fatal(err)
	}
	sib0ExpectedBalance = NewBalance(0, 100+4*(walletAtomMultiplier*4+10))
	if sib0Wallet.Balance.Compare(sib0ExpectedBalance) != 0 {
		t.Error("sibling did not have expected balance after compensation")
	}

	sib1Wallet, err = s.LoadWallet(4)
	if err != nil {
		t.Fatal(err)
	}
	sib1ExpectedBalance := NewBalance(0, 100+3*(walletAtomMultiplier*4+10))
	if sib1Wallet.Balance.Compare(sib1ExpectedBalance) != 0 {
		t.Error("sibling did not have expected balance after compensation")
	}

	// Run ExecuteCompensation again, which will deplete the funds of w2. Then
	// verify that w2 has been removed from the quorum.
	s.ExecuteCompensation()
	_, err = s.LoadWallet(2)
	if err == nil {
		t.Error("Was able to load a wallet that should have been deleted for insufficient balance.")
	}

	// Verify that siblings are not compensated for the wallet that got deleted.
	sib0Wallet, err = s.LoadWallet(3)
	if err != nil {
		t.Fatal(err)
	}
	sib0ExpectedBalance = NewBalance(0, 100+4*(walletAtomMultiplier*4+10)+3*(walletAtomMultiplier*3))
	if sib0Wallet.Balance.Compare(sib0ExpectedBalance) != 0 {
		t.Error("sibling did not have expected balance after compensation when a wallet was deleted")
	}
}
