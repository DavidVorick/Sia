package participant

import (
	"fmt"
	"quorum"
	"siacrypto"
	"time"
)

const (
	StepDuration time.Duration = 800 * time.Millisecond
)

// compile() takes the list of heartbeats and uses them to advance the state.
//
// Needs updated error handling
func (p *Participant) compile() {
	// Lock down s.heartbeats and quorum for editing
	p.heartbeatsLock.Lock()

	// fetch a sibling ordering
	siblingOrdering := p.quorum.SiblingOrdering()

	// Read heartbeats, process them, then archive them.
	for _, i := range siblingOrdering {
		// each sibling must submit exactly 1 heartbeat
		if len(p.heartbeats[i]) != 1 {
			fmt.Printf("Tossing sibling %v for %v heartbeats\n", i, len(p.heartbeats[i]))
			p.quorum.TossSibling(i)
			continue
		}

		// this is the only way I know to access the only element of a map;
		// the key is unknown
		fmt.Println("Confirming Sibling", i)
		for _, hb := range p.heartbeats[i] {
			p.quorum.IntegrateSiblingEntropy(hb.entropy)
			// process all scripts within heartbeat
		}

		// clear heartbeat list for next block
		p.heartbeats[i] = make(map[siacrypto.TruncatedHash]*heartbeat)
	}

	// copy the new seed into the quorum
	p.quorum.IntegrateGerm()

	// print the status of the quorum after compiling
	fmt.Print(p.quorum.Status())

	p.heartbeatsLock.Unlock()

	// create new heartbeat (it gets broadcast automatically), if in quorum
	if p.self.Index() != 255 {
		p.newSignedHeartbeat()
	}

	return
}

// Tick() updates s.CurrentStep, and calls compile() when all steps are complete
func (p *Participant) tick() {
	p.tickingLock.Lock()
	if p.ticking {
		p.tickingLock.Unlock()
		return
	}
	p.ticking = true
	p.tickingLock.Unlock()

	// Every StepDuration, advance the state stage
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		p.stepLock.Lock()
		if p.currentStep == quorum.QuorumSize {
			fmt.Println("compiling")
			p.currentStep = 1
			p.stepLock.Unlock() // compile needs stepLock unlocked
			p.compile()
		} else {
			fmt.Println("stepping")
			p.currentStep += 1
			p.stepLock.Unlock()
		}
	}
}
