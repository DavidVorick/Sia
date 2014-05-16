package quorum

import (
	"common"
	"common/crypto"
	"fmt"
)

/*
The Bootstrapping Process
1. Announce ourselves as a participant to the bootstrap address
2. The bootstrap address finds an index for the new participant
3. The bootstrap address announces the new participant with its index to the quorum
4. Each participant adds the new participant to their state object
5. Each participant tells the new participant about themselves

Errors will happen if anyone tries to bootstrap after the first few seconds, this is not a secure procedure
*/

// Bootstrapping
var bootstrapAddress = common.Address{
	ID:   1,
	Host: "localhost",
	Port: 9988,
}

// Announce ourself to the bootstrap address, who will announce us to the quorum
func (p *Participant) joinSia() (err error) {
	m := &common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.HandleJoinSia",
		Args: *p.self,
		Resp: nil,
	}
	p.messageRouter.SendAsyncMessage(m)
	return
}

// Adds a new Sibling, and then announces them with their index
// Currently not safe - Siblings need to be added during compile()
func (p *Participant) HandleJoinSia(s Sibling, arb *struct{}) (err error) {
	// find index for Sibling
	p.quorum.siblingsLock.Lock()
	i := 0
	for i = 0; i < common.QuorumSize; i++ {
		if p.quorum.siblings[i] == nil {
			break
		}
	}
	p.quorum.siblingsLock.Unlock()
	s.index = byte(i)
	err = p.AddNewSibling(s, nil)
	if err != nil {
		return
	}

	// see if the quorum is full
	if i == common.QuorumSize {
		return fmt.Errorf("failed to add Sibling")
	}

	// now announce a new Sibling at index i
	p.broadcast(&common.Message{
		Proc: "Participant.AddNewSibling",
		Args: s,
		Resp: nil,
	})
	return
}

// Add a Sibling to the state, tell the Sibling about ourselves
func (p *Participant) AddNewSibling(s Sibling, arb *struct{}) (err error) {
	if int(s.index) > len(p.quorum.siblings) {
		err = fmt.Errorf("sibling index exceeds lenght of siblings array")
		return
	}

	p.quorum.siblingsLock.RLock()
	if p.quorum.siblings[s.index] != nil {
		p.quorum.siblingsLock.RUnlock()
		err = fmt.Errorf("sibling already exists at targeted location")
		return
	}
	p.quorum.siblingsLock.RUnlock()

	// for this sibling, make the heartbeat map and add the default heartbeat
	hb := new(heartbeat)
	p.quorum.heartbeatsLock.Lock()
	p.quorum.siblingsLock.Lock()
	p.quorum.heartbeats[s.index] = make(map[crypto.TruncatedHash]*heartbeat)

	// get the hash of the default heartbeat
	ehb, err := hb.GobEncode()
	if err != nil {
		return
	}
	hbHash, err := crypto.CalculateTruncatedHash(ehb)
	if err != nil {
		return
	}
	p.quorum.heartbeats[s.index][hbHash] = hb
	p.quorum.heartbeatsLock.Unlock()

	compare := s.compare(p.self)
	if compare == true {
		// add our self object to the correct index in Sibings
		p.self.index = s.index
		p.quorum.siblings[s.index] = p.self
		p.quorum.tickingLock.Lock()
		p.quorum.ticking = true
		p.quorum.tickingLock.Unlock()
		go p.tick()
	} else {
		// add the Sibling to siblings
		p.quorum.siblings[s.index] = &s

		// tell the new guy about ourselves
		p.messageRouter.SendAsyncMessage(&common.Message{
			Dest: s.address,
			Proc: "Participant.AddNewSibling",
			Args: *p.self,
			Resp: nil,
		})
	}
	p.quorum.siblingsLock.Unlock()
	return
}
