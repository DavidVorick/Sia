package consensus

import (
	"testing"
	"time"
)

// Ensures that Tick() updates CurrentStep
func TestRegularTick(t *testing.T) {
	// test takes StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	p := new(Participant)
	p.currentStep = 1
	go p.tick()

	// verify that tick is updating CurrentStep
	time.Sleep(StepDuration)
	time.Sleep(50 * time.Millisecond)
	p.stepLock.Lock()
	if p.currentStep != 2 {
		t.Fatal("s.currentStep failed to update correctly:", p.currentStep)
	}
	p.stepLock.Unlock()
}

// Now that compilation is a lot more complex, I'm not sure that this makes
// much sense to test
// ensures Tick() calles compile() and then resets the counter to step 1
/* func TestCompilationTick(t *testing.T) {
	// test takes StepDuration seconds; skip for short testing
	if testing.Short() {
		t.Skip()
	}

	p := new(Participant)
	p.currentStep = quorum.QuorumSize
	p.self = new(quorum.Sibling)
	go p.tick()

	// verify that tick is wrapping around properly
	time.Sleep(StepDuration)
	time.Sleep(50 * time.Millisecond)
	p.stepLock.Lock()
	if p.currentStep != 1 {
		t.Error("p.currentStep failed to roll over:", p.currentStep)
	}
	p.stepLock.Unlock()
} */
