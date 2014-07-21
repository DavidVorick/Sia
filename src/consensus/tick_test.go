package consensus

// Package uses port range 11000 - 11025

import (
	"network"
	"state"
	"testing"
	"time"
)

// TestSynchronizedTick checks that all of the required logic for
// Participant.tick() runs without error when the participant is synchronized
// to the quorum.
func TestSynchronizedTick(t *testing.T) {
	// Create a bootstrapped participant to test with.
	rpcs, err := network.NewRPCServer(11025)
	if err != nil {
		t.Fatal(err)
	}
	p, err := CreateBootstrapParticipant(rpcs, "../../filesCreatedDuringTesting/TestSynchronizedTick", 24)
	if err != nil {
		t.Fatal(err)
	}

	// Check that current step is initialized to 1.
	p.currentStepLock.RLock()
	if p.currentStep != 1 {
		t.Error("p.currentStep not initializing to 1")
	}
	p.currentStepLock.RUnlock()

	// Sleep for 1 step and see if current step has increased.
	time.Sleep(25 * time.Millisecond)
	time.Sleep(StepDuration)
	p.currentStepLock.RLock()
	if p.currentStep != 2 {
		t.Error("p.currentStep is not incrementing correctly each StepDuration")
	}
	p.currentStepLock.RUnlock()

	// Check that the quorum height has been initialized to 0.
	if p.engine.Metadata().Height != 0 {
		t.Error("Quorum height not initialized to 0")
	}

	// Set the currentStep to trigger a compile and wait for the compile to
	// trigger.
	p.currentStepLock.Lock()
	p.currentStep = state.QuorumSize
	p.currentStepLock.Unlock()
	time.Sleep(StepDuration)

	// Check that the height of the quorum has increased.
	p.engineLock.RLock()
	if p.engine.Metadata().Height != 1 {
		t.Error("Quorum height has not increased to 1 after compilation")
	}
	p.engineLock.RUnlock()

	// Is there some way to check that a new heartbeat was created and broadcast
	// to the newtork???
}
