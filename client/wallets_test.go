package client

import (
	"bytes"
	"os"
	"siacrypto"
	"state"
	"testing"
)

// Two things. First, this nees to be better commented. Second, there should be
// more spaces, breaking the tests into logical units are are 1 thought/test
// each. Usually this means 3-5 lines, though it can be more or less.
func TestSaveLoadWallet(t *testing.T) {
	var id1, id2 state.WalletID
	var keypair1, keypair2 Keypair
	id1 = state.WalletID(siacrypto.RandomUint64())
	var err error
	keypair1.PK, keypair1.SK, err = siacrypto.CreateKeyPair()
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
	if !bytes.Equal(keypair1.PK[:], keypair2.PK[:]) {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	if !bytes.Equal(keypair1.SK[:], keypair2.SK[:]) {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	os.Remove("tempFile")
}
