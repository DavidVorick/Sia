package consensus

import (
	"testing"
	"time"

	"github.com/NebulousLabs/Sia/state"
)

// TestSynchronizedTick checks that all of the required logic for
// Participant.tick() runs without error when the participant is synchronized
// to the quorum.
func TestSynchronizedTick(t *testing.T) {
	var p Participant
	go p.tick()

	// Sleep for 1 step and see if current step has increased.
	time.Sleep(StepDuration + 25*time.Millisecond)
	p.tickLock.RLock()
	if p.currentStep != 1 {
		t.Error("p.currentStep is not incrementing correctly each StepDuration")
	}
	p.tickLock.RUnlock()

	// Check that the quorum height has been initialized to 0.
	p.engineLock.RLock()
	if p.engine.Metadata().Height != 0 {
		t.Error("Quorum height not initialized to 0")
	}
	p.engineLock.RUnlock()

	// Set the currentStep to trigger a compile and wait for the compile to
	// trigger.
	p.tickLock.RLock()
	p.currentStep = state.QuorumSize
	p.tickLock.RUnlock()
	time.Sleep(StepDuration)

	// Check that the height of the quorum has increased.
	p.engineLock.RLock()
	if p.engine.Metadata().Height != 1 {
		t.Error("Quorum height has not increased to 1 after compilation")
	}
	p.engineLock.RUnlock()

	// Is there some way to check that a new heartbeat was created and
	// broadcast to the newtork???
}
