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
	// Create a ticker that will pulse every StepDuration
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		// Once cryptographic synchronization is implemented, there
		// will be an additional sleep placed here for some volume of
		// seconds that will keep the participant synchronized to a
		// much higher degree of accuracy.

		p.currentStepLock.Lock()
		if p.currentStep == state.QuorumSize {
			p.currentStep = 0

			// Have the engine compile and integrate the block.
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
