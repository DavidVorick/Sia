package consensus

import (
	"delta"
	"siacrypto"
	"state"
	"time"
)

const (
	StepDuration = 1800 * time.Millisecond
)

// condenseBlock assumes that a heartbeat has a valid signature and that the
// parent is the correct parent.
func (p *Participant) condenseBlock() (b delta.Block) {
	// Lock the engine and the updates variables
	p.engineLock.RLock()
	defer p.engineLock.RUnlock()

	p.updatesLock.Lock()
	defer p.updatesLock.Unlock()

	// Set the height and parent of the block.
	b.Height = p.engine.Metadata().Height
	b.ParentBlock = p.engine.Metadata().ParentBlock

	// Take each update and condense them into a single non-repetitive block.
	for i := range p.updates {
		if len(p.updates[i]) == 1 {
			for _, u := range p.updates[i] {
				// Add the heartbeat
				b.Heartbeats[i] = u.Heartbeat
				b.HeartbeatSignatures[i] = u.HeartbeatSignature

				// Add the other stuff (tbi)
			}
		}
		p.updates[i] = make(map[siacrypto.Hash]Update) // clear map for next cycle
	}
	return
}

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
				p.engineLock.Lock()
				p.engine.Compile(b)
				p.engineLock.Unlock()
			} else {
				// p.appendBlock(b)
			}

			// Broadcast a new update to the quorum.
			p.newSignedUpdate()

			p.currentStep = 1
		} else {
			p.currentStep += 1
		}
		p.currentStepLock.Unlock()
	}
}
