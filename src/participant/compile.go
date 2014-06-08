package participant

import (
	"fmt"
	"quorum/script"
	"siacrypto"
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

		// this is the only way I know to access the only element of a map; the key
		// is unknown
		fmt.Println("Confirming Sibling", i)
		for _, hb := range p.heartbeats[i] {
			p.quorum.IntegrateSiblingEntropy(hb.entropy)
			for _, si := range hb.scriptInputs {
				block := p.quorum.LoadScript(si.WalletID)
				s := script.Script{block}
				s.Execute(si.Input, &p.quorum)
			}
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
	if p.self.Index() == 255 {
		siblings := p.quorum.Siblings()
		for _, sibling := range siblings {
			if sibling.Compare(p.self) {
				p.self = sibling
			}
		}
	}
	if p.self.Index() != 255 {
		p.newSignedHeartbeat()
	}

	return
}
