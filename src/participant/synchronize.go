package participant

import (
	"quorum"
	"siacrypto"
)

type Synchronize struct {
	currentStep int
	heartbeats  [quorum.QuorumSize]map[siacrypto.TruncatedHash]*heartbeat
}

func (s *Synchronize) GobEncode() (gobSynchronize []byte, err error) {
	if s == nil {
		s = new(Synchronize)
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(s.currentStep)
	if err != nil {
		return
	}
	err = encoder.Encode(s.heartbeats)
	if err != nil {
		return
	}
	gobSynchronize = w.Bytes()
	return
}

func (s *Synchronize) GobDecode(gobSynchronize []byte) (err error) {
	if s == nil {
		s = new(Synchronize)
	}

	r := bytes.NewBuffer(gobSynchronize)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&s.currentStep)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.heartbeats)
	if err != nil {
		return
	}
	return
}

// Synchronize just sends over participant.currentStep, and is a temporary
// function. It's not secure or trusted and is highely exploitable.
//
// This needs to be changed so that someone requests a synchronize, and a rv
// is sent
func (p *Participant) Synchronize(arb struct{}, s *Synchronize) (err error) {
	p.stepLock.Lock()
	s.currentStep = p.currentStep
	p.stepLock.Unlock()

	p.heartbeatsLock.Lock()
	s.heartbeats = p.heartbeats
	p.heartbeatsLock.Unlock()
	return
}
