package client

import (
	"math/rand"
	"os"
	"quorum"
	"siacrypto"
	"testing"
)

func TestSaveLoadWallet(t *testing.T) {
	var (
		id1, id2           quorum.WalletID
		keypair1, keypair2 *siacrypto.Keypair
		err                error
	)
	keypair1 = new(siacrypto.Keypair)
	keypair2 = new(siacrypto.Keypair)
	id1 = quorum.WalletID(uint64(rand.Uint32()) * uint64(rand.Uint32())) //There isn't a Uint64()
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
	if keypair1.PK.Compare(keypair2.PK) == false {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	if keypair1.SK.Compare(keypair2.SK) == false {
		t.Fatal("Wallet's key pair not preserved when loaded")
	}
	_ = os.Remove("tempFile")
}
