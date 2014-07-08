package quorum

import (
	"siacrypto"
	"testing"
)

func TestSnapshotHeaderEncoding(t *testing.T) {
	var sh snapshotHeader
	sh.walletLookupOffset = uint32(siacrypto.RandomUint64())
	sh.wallets = uint32(siacrypto.RandomUint64())

	var dsh snapshotHeader
	esh, err := sh.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	err = dsh.GobDecode(esh)
	if err != nil {
		t.Fatal(err)
	}

	if sh.walletLookupOffset != dsh.walletLookupOffset {
		t.Error("walletLookupOffset does not match")
	}
	if sh.wallets != dsh.wallets {
		t.Error("wallets does not match")
	}
}
