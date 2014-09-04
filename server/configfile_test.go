package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// Two things. First, this nees to be better commented. Second, there should be
// more spaces, breaking the tests into logical units are are 1 thought/test
// each. Usually this means 3-5 lines, though it can be more or less.
func TestSaveLoadWallet(t *testing.T) {
	var id1, id2 state.WalletID
	var keypair1, keypair2 GenericWallet
	id1 = state.WalletID(siacrypto.RandomUint64())
	var err error
	keypair1.PublicKey, keypair1.SecretKey, err = siacrypto.CreateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	err = SaveWallet(id1, keypair1, "tempFile")
	if err != nil {
		t.Fatal(err)
	}
	id2, keypair2, err = LoadWallet("tempFile")
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Fatal("Wallet ID not preserved when loaded")
	}
	if !bytes.Equal(keypair1.PublicKey[:], keypair2.PublicKey[:]) {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	if !bytes.Equal(keypair1.SecretKey[:], keypair2.SecretKey[:]) {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	os.Remove("tempFile")
}
