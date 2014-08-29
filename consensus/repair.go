package consensus

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

func (p *Participant) recoverSegment(id state.WalletID) (err error) {
	// Get the wallet so that we know what we are downloading.
	w, err := p.engine.Wallet(id)
	if err != nil {
		return
	}
	if w.Sector.Atoms == 0 {
		err = errors.New("wallet has no sector")
		return
	}

	// Download K pieces into K readers, 'segments'.
	var segments []io.Reader
	var indices []byte
	for i := range segments {
		var segment []byte
		err2 := p.router.SendMessage(network.Message{
			Dest: p.engine.Metadata().Siblings[i].Address,
			Proc: "Participant.DownloadSegment",
			Args: id,
			Resp: &segment,
		})
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
	var buffer bytes.Buffer
	var writer io.Writer
	writer = &buffer
	_, err = state.RSRecover(segments, indices, writer, int(w.Sector.K))
	if err != nil {
		return
	}

	// Turn the writer into a reader.
	originalBytes := buffer.Bytes()
	reader := bytes.NewReader(originalBytes)

	// Use the writer to create the full set of segments, including the one
	// we need.
	var fullSegments [state.QuorumSize]io.Writer
	var fullSegmentBytes [state.QuorumSize]bytes.Buffer
	for i := range fullSegments {
		fullSegments[i] = &fullSegmentBytes[i]
	}
	atoms, err := state.RSEncode(reader, fullSegments, int(w.Sector.K))
	if err != nil {
		return
	}

	// Take the original segment and get its hash.
	segmentBytes := fullSegmentBytes[p.siblingIndex].Bytes()
	segmentBytesReader := bytes.NewReader(segmentBytes)
	hash, err := state.MerkleCollapse(segmentBytesReader, atoms)
	if err != nil {
		return
	}

	// Reload the wallet (timing), verify the hash, and write to disk.
	// ALSO DO A WALLET LOCK.
	w, err = p.engine.Wallet(id)
	if err != nil {
		return
	}

	if hash != w.Sector.HashSet[p.siblingIndex] {
		err = errors.New("will not recover file - hash incorrect!")
		return
	}

	file, err := os.Create(p.engine.SegmentFilename(id))
	if err != nil {
		return
	}
	file.Write(segmentBytes)
	file.Close()

	return
}

func (p *Participant) recoveryListen() {
	repairChan := p.engine.RepairChan()

	for repairRequest := range repairChan {
		go p.recoverSegment(repairRequest)
	}
}
