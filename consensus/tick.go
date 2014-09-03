package consensus

import (
	"fmt"
	"time"

	"github.com/NebulousLabs/Sia/state"
)

const (
	// StepDuration is the amount of time between each step.
	// Each block is compiled after state.QuorumSize steps.
	StepDuration = 600 * time.Millisecond
)

func (p *Participant) tick() {
	// Verify that tick() has not already been called.
	p.tickLock.Lock()
	if p.ticking {
		p.tickLock.Unlock()
		return
	}
	p.ticking = true
	p.updateStop.Unlock()

	// Create a ticker that will pulse every StepDuration
	p.tickStart = time.Now()
	ticker := time.Tick(StepDuration)
	p.tickLock.Unlock() // Unlock the mutex before entering the tick loop.
	for _ = range ticker {
		// Once cryptographic synchronization is implemented, there
		// will be an additional sleep placed here for some volume of
		// seconds that will keep the participant synchronized to a
		// much higher degree of accuracy.

		p.tickLock.Lock()
		if p.currentStep == state.QuorumSize {
			p.currentStep = 0
			p.tickLock.Unlock()

			// Have the engine condense and integrate the block,
			// then send a new heartbeat. This is done in a
			// separate thread so that the timing is not disrupted
			// in the parent thread.
			go func() {
				// Condense the list of updates into a block.
				block := p.condenseBlock()

				// Compile the block.
				p.engineLock.Lock()
				err := p.engine.Compile(block)
				p.engineLock.Unlock()
				if err != nil {
					fmt.Println(err)
				}

				// Broadcast a new update to the quorum.
				p.newSignedUpdate()
			}()
		} else {
			p.currentStep++
			p.tickLock.Unlock()
		}
	}
}
