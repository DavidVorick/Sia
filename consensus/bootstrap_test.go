package consensus

import (
	"testing"
	"time"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

func TestCreateParticipantFunctions(t *testing.T) {
	rpcs, err := network.NewRPCServer(11000)
	if err != nil {
		t.Fatal(err)
	}

	walletID := state.WalletID(24)
	p, err := CreateBootstrapParticipant(rpcs, siafiles.TempFilename("TestCreateParticipantFunctions"), walletID)
	if err != nil {
		t.Fatal(err)
	}

	var metadata state.Metadata
	p.Metadata(struct{}{}, &metadata)
	if !metadata.Siblings[0].Active() {
		t.Error("Sibling in the bootstrap position not marked as active!")
	}
	p.tickLock.Lock()
	if p.currentStep != 1 {
		t.Error("p.currentStep needs to be initialized to 1")
	}
	p.tickLock.Unlock()

	var walletIDs []state.WalletID
	p.WalletIDs(struct{}{}, &walletIDs)
	if len(walletIDs) != 2 {
		t.Error("Incorrect number of wallets returned, expeting 2:", len(walletIDs))
	}

	if testing.Short() {
		t.Skip()
	}

	time.Sleep(time.Millisecond * 50)
	time.Sleep(StepDuration)
	p.tickLock.Lock()
	if p.currentStep != 2 {
		t.Error("step counter is not increasing after bootstrap.")
	}
	p.tickLock.Unlock()
}
