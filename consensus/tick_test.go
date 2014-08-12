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
	if p.currentStep != 1 {
		t.Error("p.currentStep is not incrementing correctly each StepDuration")
	}

	// Check that the quorum height has been initialized to 0.
	if p.engine.Metadata().Height != 0 {
		t.Error("Quorum height not initialized to 0")
	}

	// Set the currentStep to trigger a compile and wait for the compile to
	// trigger.
	p.currentStep = state.QuorumSize
	time.Sleep(StepDuration)

	// Check that the height of the quorum has increased.
	if p.engine.Metadata().Height != 1 {
		t.Error("Quorum height has not increased to 1 after compilation")
	}

	// Is there some way to check that a new heartbeat was created and
	// broadcast to the newtork???
}
