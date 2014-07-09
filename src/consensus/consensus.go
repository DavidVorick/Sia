package consensus

// condenseBlock assumes that a heartbeat has a valid signature and that the
// parent is the correct parent.
func (p *Participant) condenseBlock() (b *block) {
	b = new(block)
	b.height = p.quorum.Height()
	b.parent = p.quorum.Parent()

	p.heartbeatsLock.Lock()
	for i := range p.heartbeats {
		fmt.Printf("Sibling %v: %v heartbeats\n", i, len(p.heartbeats[i]))
		if len(p.heartbeats[i]) == 1 {
			// the map has only one element, but the key is unknown
			for _, hb := range p.heartbeats[i] {
				b.heartbeats[i] = hb // place heartbeat into block, if valid
			}
		}
		p.heartbeats[i] = make(map[siacrypto.Hash]*heartbeat) // clear map for next cycle
	}
	p.heartbeatsLock.Unlock()
	return
}
