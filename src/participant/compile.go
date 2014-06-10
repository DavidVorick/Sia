package participant

import (
	"fmt"
	"quorum/script"
	"siacrypto"
)

// compile() is a messy, messy function that takes the quorum from one point to
// the next. I'm brainstorming ways to clean it up, because there's not that
// much that needs to happen and not much complexity to how it needs to happen.
// It's just one of those cases where it comes out in english a lot smoother
// than comes out in code.
func (p *Participant) compile() {
	// fill out the basic information for the new block
	var b block
	b.height = p.currentBlock
	p.currentBlock += 1
	b.parent = p.previousBlock

	p.heartbeatsLock.Lock()
	siblingOrdering := p.quorum.SiblingOrdering()
	for _, i := range siblingOrdering {
		if len(p.heartbeats[i]) != 1 {
			fmt.Printf("Tossing sibling %v for %v heartbeats\n", i, len(p.heartbeats[i]))
			p.quorum.TossSibling(i)
			continue
		}

		// this is the only way I know to access the only element of a map; the key
		// is unknown
		fmt.Printf("Confirming Sibling %v\n", i)
		for _, hb := range p.heartbeats[i] {
			b.heartbeats[i] = hb // add heartbeat to block
			p.quorum.IntegrateSiblingEntropy(hb.entropy)
			for _, si := range hb.scriptInputs {
				scriptBlock := p.quorum.LoadScriptBlock(si.WalletID)
				if scriptBlock == nil {
					continue
				}
				s := script.Script{scriptBlock}
				s.Execute(si.Input, &p.quorum)

				// will soon be replaced with a single line: p.quorum.HandleScript(si)
			}
		}

		p.heartbeats[i] = make(map[siacrypto.Hash]*heartbeat)
	}

	// save the block
	p.saveBlock(&b)

	p.quorum.IntegrateGerm()     // cycles the entropy
	fmt.Print(p.quorum.Status()) // helps with debugging
	p.heartbeatsLock.Unlock()

	// if not a sibling, check to see if you've been added as a sibling. This is
	// a crude way of doing it but it gets the job done.
	if p.self.Index() == 255 {
		siblings := p.quorum.Siblings()
		for _, sibling := range siblings {
			if sibling.Compare(p.self) {
				p.self = sibling
			}
		}
	}

	// only send a heartbeat if you are a sibling
	if p.self.Index() != 255 {
		p.newSignedHeartbeat()
	}

	return
}
