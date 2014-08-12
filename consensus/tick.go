package consensus

import (
	"time"

	"github.com/NebulousLabs/Sia/state"
)

const (
	// StepDuration is the amount of time between each step.
	// Each block is compiled after state.QuorumSize steps.
	StepDuration = 1800 * time.Millisecond
)

// The final step is a block compilation step. At this point,

// tick maintains the consensus rhythm for the participant, updating the
// counter that tells the participant which Updates and Blocks are acceptable,
// which are early, and which are late.
func (p *Participant) tick() {
	// Verify that tick() has not already been called.
	if p.ticking {
		return
	}
	p.ticking = true

	// Set the step to '0', and assume that tick was called at a time where
	// the rest of the network has also reset to step 0.
	p.currentStepLock.Lock()
	p.currentStep = 0
	p.currentStepLock.Unlock()

	// Create a ticker that will pulse every StepDuration
	p.tickStart = time.Now()
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		// Once cryptographic synchronization is implemented, there
		// will be an additional sleep placed here for some volume of
		// seconds that will keep the participant synchronized to a
		// much higher degree of accuracy.

		p.currentStepLock.Lock()
		if p.currentStep == state.QuorumSize {
			p.currentStep = 0

			// Have the engine condense and integrate the block,
			// then send a new heartbeat. This is done in a
			// separate thread so that the timing is not disrupted
			// in the parent thread.
			go func() {
				// First condense the block, then set the current step to 1. The order
				// shouldn't matter because currentStep is locked by a mutex.
				block := p.condenseBlock()

				p.engineLock.Lock()
				p.engine.Compile(block)
				p.engineLock.Unlock()

				// Broadcast a new update to the quorum.
				p.newSignedUpdate()
			}()
		} else {
			p.currentStep++
		}
		p.currentStepLock.Unlock()
	}
}
