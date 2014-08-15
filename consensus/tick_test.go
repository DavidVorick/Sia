package consensus

import (
	"testing"
	"time"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

// TestSynchronizedTick checks that all of the required logic for
// Participant.tick() runs without error when the participant is synchronized
// to the quorum.
func TestSynchronizedTick(t *testing.T) {
	mr, err := network.NewRPCServer(11300)
	if err != nil {
		t.Fatal(err)
	}
	p, err := CreateBootstrapParticipant(mr, siafiles.TempFilename("TestSynchronizedTick"), 1)
	if err != nil {
		t.Fatal(err)
	}
	go p.tick()

	// Sleep for 1 step and see if current step has increased.
	time.Sleep(StepDuration + 25*time.Millisecond)
	p.tickLock.RLock()
	if p.currentStep != 2 {
		t.Error("p.currentStep is not incrementing correctly each StepDuration")
	}
	p.tickLock.RUnlock()

	// Set the currentStep to trigger a compile and wait for the compile to
	// trigger.
	p.tickLock.Lock()
	p.currentStep = state.QuorumSize
	p.tickLock.Unlock()
	time.Sleep(StepDuration)

	// Check that the height of the quorum has increased.
	p.engineLock.RLock()
	if p.engine.Metadata().Height != 2 {
		t.Error("Quorum height has not increased to 2 after compilation")
	}
	p.engineLock.RUnlock()
}
