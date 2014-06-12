package participant

func (p *Participant) appendBlock(b *block) {
	if p.recentBlocks[b.height] != nil {
		return
	}

	p.recentBlocks[b.height] = b

	// if we are at a certain block value, create a snapshot?
	// not worried about that for now
}
