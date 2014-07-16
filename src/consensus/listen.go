package consensus

/* import (
	"quorum"
	"time"
)

const (
	StepDuration time.Duration = 800 * time.Millisecond
)

func (p *Participant) tick() {
	ticker := time.Tick(StepDuration)
	for _ = range ticker {
		p.stepLock.Lock()
		if p.currentStep == int(quorum.QuorumSize) {
			b := p.condenseBlock()

			//p.appendBlock(b)

			if p.synchronized {
				p.compile(b)
			}

			p.currentStep = 1
		} else {
			p.currentStep += 1
		}
		p.stepLock.Unlock()
	}
} */
