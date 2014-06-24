package client

import (
	"math/rand"
	"os"
	"testing"
)

func TestSaveLoadWallet(t *testing.T) {
	var wal0, wal1 *Wallet
	wal0 = new(Wallet)
	wal1 = new(Wallet)
	wal0.ID = rand.Uint32()
	wal0.Type = "Generic"
	err := SaveWallet(wal0.ID, wal0.Type, "tempFile")
	if err != nil {
		t.Fatal(err)
	}
	wal1, err = LoadWallet("tempFile")
	if err != nil {
		panic(err)
	}
	if wal0.ID != wal1.ID {
		t.Fatal("Wallet ID not preserved when loaded")
	}
	if wal0.Type != wal1.Type {
		t.Fatal("Wallet Type not preserved when loaded")
	}
	err = os.Remove("tempFile")
	if err != nil {
		t.Fatal(err)
	}
}
