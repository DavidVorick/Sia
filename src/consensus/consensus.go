package consensus

// the blockHistoryHeader is the header that preceeds the block history file,
// containing an index of all the blocks in the history and their data offsets
// in the file.
type blockHistoryHeader struct {
	latestBlock  uint32
	blockOffsets [SnapshotLen]uint32
}

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

func (bhh *blockHistoryHeader) GobEncode() (gobBHH []byte, err error) {
	gobBHH = make([]byte, BlockHistoryHeaderSize)
	encodedInt := siaencoding.EncUint32(bhh.latestBlock)
	copy(gobBHH, encodedInt[:])
	offset := len(encodedInt)

	for i := range bhh.blockOffsets {
		encodedInt = siaencoding.EncUint32(bhh.blockOffsets[i])
		copy(gobBHH[offset:], encodedInt[:])
		offset += len(encodedInt)
	}

	return
}

func (bhh *blockHistoryHeader) GobDecode(gobBHH []byte) (err error) {
	if len(gobBHH) != BlockHistoryHeaderSize {
		err = fmt.Errorf("gobBHH has wrong size, cannot decode!")
	}

	bhh.latestBlock = siaencoding.DecUint32(gobBHH[:4])
	offset := 4

	for i := range bhh.blockOffsets {
		bhh.blockOffsets[i] = siaencoding.DecUint32(gobBHH[offset : offset+4])
		offset += 4
	}

	return
}
