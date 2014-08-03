package consensus

import (
	"state"
	"time"
)

const (
	// StepDuration is the amount of time between each step.
	// Each block is compiled after state.QuorumSize steps.
	StepDuration = 1800 * time.Millisecond
)

// TODO: add docstring
func (p *Participant) tick() {
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		p.currentStepLock.Lock()
		if p.currentStep == state.QuorumSize {
			// First condense the block, then set the current step to 1. The order
			// shouldn't matter because currentStep is locked by a mutex.
			b := p.condenseBlock()
			p.currentStep = 1

			// If synchronized, give the block to the engine for processing.
			// Otherwise, save the block in a map that is used to assist
			// synchronization.
			if p.synchronized {
				p.engineLock.Lock()
				p.engine.Compile(b)
				p.engineLock.Unlock()

				// Broadcast a new update to the quorum.
				p.newSignedUpdate()
			} else {
				// p.appendBlock(b)
			}
		} else {
			p.currentStep++
		}
		p.currentStepLock.Unlock()
	}
}
