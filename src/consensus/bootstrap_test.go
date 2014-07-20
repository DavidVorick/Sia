package consensus

import (
	"network"
	"state"
	"testing"
)

func TestCreateParticipantFunctions(t *testing.T) {
	rpcs, err := network.NewRPCServer(11000)
	if err != nil {
		t.Fatal(err)
	}

	walletID := state.WalletID(24)
	p, err := CreateBootstrapParticipant(rpcs, "../../filesCreatedDuringTesting/TestCreateParticipantFunctions", walletID)
	if err != nil {
		t.Fatal(err)
	}

	var metadata state.StateMetadata
	p.Metadata(struct{}{}, &metadata)
	if !metadata.Siblings[0].Active {
		t.Error("No sibling in the bootstrap position.")
	}

	var walletIDs []state.WalletID
	p.GetWallets(struct{}{}, &walletIDs)
	if len(walletIDs) != 2 {
		t.Error("Incorrect number of wallets returned, expeting 2:", len(walletIDs))
	}
}
