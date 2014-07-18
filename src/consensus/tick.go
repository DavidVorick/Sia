package consensus

import (
	"state"
	"time"
)

const (
	StepDuration = 800 * time.Millisecond
)

func (p *Participant) tick() {
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		p.currentStepLock.Lock()
		if p.currentStep == state.QuorumSize {
			b := p.condenseBlock()

			// If synnchronized, give the block to the engine for processing.
			// Otherwise, save the block in a map that is used to assist
			// synchronization.
			if p.synchronized {
				p.engine.Compile(b)
			} else {
				// p.appendBlock(b)
			}

			p.currentStep = 1
		} else {
			p.currentStep += 1
		}
		p.currentStepLock.Unlock()
	}
}
