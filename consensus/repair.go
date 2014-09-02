package consensus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

func (p *Participant) recoverSegment(id state.WalletID) (err error) {
	// Get the wallet so that we know what we are downloading.
	p.engineLock.RLock()
	w, err := p.engine.Wallet(id)
	if err != nil {
		return
	}
	p.engineLock.RUnlock()

	if w.Sector.Atoms == 0 {
		err = fmt.Errorf("wallet %v has no sector", id)
		return
	}

	// Download K pieces into K readers, 'segments'.
	var segments []io.Reader
	var indices []byte
	for i := range p.engine.Metadata().Siblings {
		p.engineLock.RLock()
		var segment []byte
		err2 := p.router.SendMessage(network.Message{
			Dest: p.engine.Metadata().Siblings[i].Address,
			Proc: "Participant.DownloadSegment",
			Args: id,
			Resp: &segment,
		})
		p.engineLock.RUnlock()
		if err2 != nil {
			continue
		}

		segments = append(segments, bytes.NewReader(segment))
		indices = append(indices, byte(i))

		if len(segments) >= int(w.Sector.K) {
			break
		}
	}

	// Check that enough pieces were gathered.
	if len(segments) < int(w.Sector.K) {
		err = errors.New("could not repair segment")
		return
	}

	// Have the state decode the segments into a new sector.
	buffer := new(bytes.Buffer)
	_, err = state.RSRecover(segments, indices, buffer, int(w.Sector.K))
	if err != nil {
		return
	}

	// Use the writer to create the full set of segments, including the one
	// we need.
	fullSegments := make([][]byte, state.QuorumSize)
	atoms, err := state.RSEncode(buffer, fullSegments, int(w.Sector.K))
	if err != nil {
		return
	}

	// Take the original segment and get its hash.
	p.engineLock.RLock()
	segment := fullSegments[p.engine.SiblingIndex()]
	hash, err := state.MerkleCollapse(bytes.NewReader(segment), atoms)
	if err != nil {
		return
	}
	p.engineLock.RUnlock()

	// Reload the wallet (timing), verify the hash, and write to disk.
	p.engineLock.RLock()
	w, err = p.engine.Wallet(id)
	if err != nil {
		return
	}
	p.engineLock.RUnlock()

	if hash != w.Sector.HashSet[p.engine.SiblingIndex()] {
		err = errors.New("will not recover file - hash incorrect!")
		return
	}

	file, err := os.Create(p.engine.SegmentFilename(id))
	if err != nil {
		return
	}
	file.Write(segment)
	file.Close()

	return
}

func (p *Participant) recoveryListen() {
	repairChan := p.engine.RepairChan()

	for repairRequest := range repairChan {
		go p.recoverSegment(repairRequest)
	}
}
