package participant

import (
	"fmt"
	"quorum"
	"time"
)

const (
	StepDuration time.Duration = 3 * time.Second
)

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
