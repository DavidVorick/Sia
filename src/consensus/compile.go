package consensus

import (
	"delta"
	//"fmt"
)

// compile() is a messy, messy function that takes the quorum from one point to
// the next. I'm brainstorming ways to clean it up, because there's not that
// much that needs to happen and not much complexity to how it needs to happen.
// It's just one of those cases where it comes out in english a lot smoother
// than comes out in code.
func (p *Participant) compile(b *delta.Block) {
	/* siblingOrdering := p.quorum.SiblingOrdering()
	for _, i := range siblingOrdering {
		if b.heartbeats[i] == nil {
			fmt.Printf("Tossing sibling %v for %v heartbeats\n", i, len(p.heartbeats[i]))
			p.quorum.TossSibling(i)
			continue
		}

		for j := range b.heartbeats[i].uploadAdvancements {
			p.quorum.AdvanceUpload(&b.heartbeats[i].uploadAdvancements[j])
		}

		fmt.Printf("Confirming Sibling %v\n", i)
		p.quorum.IntegrateSiblingEntropy(b.heartbeats[i].entropy)
		for _, si := range b.heartbeats[i].scriptInputs {
			si.Execute(&p.quorum)
		}
	}
	p.quorum.ExecuteCompensation()
	p.quorum.IntegrateGerm()
	p.quorum.ProcessEvents()
	p.quorum.AdvanceBlock(b.parent)
	p.engine.SaveBlock(b)
	fmt.Print(p.quorum.Status())

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
	*/
	return
}
