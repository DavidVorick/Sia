package client

import (
	"bytes"
	"os"
	"siacrypto"
	"state"
	"testing"
)

func TestSaveLoadWallet(t *testing.T) {
	var id1, id2 state.WalletID
	keypair1, keypair2 := new(Keypair), new(Keypair)
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
