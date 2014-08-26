package consensus

import (
	"time"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/state"
)

// AddScriptInput is an RPC that appends a script input to
// Participant.scriptInputs.
func (p *Participant) AddScriptInput(si state.ScriptInput, _ *struct{}) (err error) {
	p.updatesLock.Lock()
	p.scriptInputs = append(p.scriptInputs, si)
	p.updatesLock.Unlock()
	return
}

// Block is an RPC that returns a block of a specific height. Participants only
// keep a history of so many blocks, so asking for future blocks or expired
// blocks will return an error.
func (p *Participant) Block(blockHeight uint32, block *delta.Block) (err error) {
	p.engineLock.RLock()
	*block, err = p.engine.LoadBlock(blockHeight)
	p.engineLock.RUnlock()
	return
}

// ConsensusProgressStruct is a struct that indicates how far progressed through the
// current round of consensus the current participant is.
type ConsensusProgressStruct struct {
	Height              uint32
	CurrentStep         byte
	CurrentStepProgress time.Duration
}

// ConsensusProgress is an RPC that returns the progress of the participant and
// quorum through the current round of consensus. It is useful for indicating
// when the next block will be ready.
func (p *Participant) ConsensusProgress(_ struct{}, cps *ConsensusProgressStruct) (err error) {
	p.tickLock.RLock()
	p.engineLock.RLock()

	cps.Height = p.engine.Metadata().Height
	cps.CurrentStep = p.currentStep
	cps.CurrentStepProgress = time.Since(p.tickStart) % StepDuration

	p.tickLock.RUnlock()
	p.engineLock.RUnlock()
	return
}

func (p *Participant) DownloadSector(id state.WalletID, sector *[]byte) (err error) {
	p.engineLock.RLock()
	*sector, err = p.engine.DownloadSector(id)
	p.engineLock.RUnlock()
	return
}

// Metadata is an RPC that returns the current state metadata.
func (p *Participant) Metadata(_ struct{}, smd *state.Metadata) (err error) {
	p.engineLock.RLock()
	*smd = p.engine.Metadata()
	p.engineLock.RUnlock()
	return
}

// UploadSegment accepts a SegmentUpload contianing a wallet id, an update
// index, and a new segment. This is processed by the engine. If the
// segmentupload is accepted, then an update advancement is added to be sent to
// the quorum in the next heartbeat.
func (p *Participant) UploadSegment(upload delta.SegmentUpload, accepted *bool) (err error) {
	p.engineLock.Lock()
	*accepted, err = p.engine.ProcessSegmentUpload(upload)
	p.engineLock.Unlock()
	if err != nil {
		return
	}

	if *accepted {
		// Add an upload advancement confirming that we have our
		// segment of this upload.
		newAdvancement := state.UpdateAdvancement{
			SiblingIndex: p.siblingIndex,
			WalletID:     upload.WalletID,
			UpdateIndex:  upload.UpdateIndex,
		}
		p.updatesLock.Lock()
		p.updateAdvancements = append(p.updateAdvancements, newAdvancement)
		p.updatesLock.Unlock()
	}

	return
}

/*
// UpdateSegment is an RPC that allows hosts to submit diffs that match updates
// that have been confirmed by consensus.
func (p *Participant) UpdateSegment(sd delta.SegmentDiff, accepted *bool) (err error) {
	p.engineLock.Lock()
	*accepted, err = p.engine.UpdateSegment(sd)
	p.engineLock.Unlock()

	if *accepted {
		// Submit a notification to the quorum that a match has been uploaded.
		newAdvancement := state.UpdateAdvancement{
			Index:    p.engine.SiblingIndex(),
			UpdateID: sd.UpdateID,
		}
		p.updatesLock.Lock()
		p.updateAdvancements = append(p.updateAdvancements, newAdvancement)
		p.updatesLock.Unlock()
	}

	return
}
*/

func (p *Participant) Wallet(id state.WalletID, w *state.Wallet) (err error) {
	*w, err = p.engine.Wallet(id)
	return
}

// Not sure what the use is for this, mostly wallets are downloaded via
// snapshots. Doesn't hurt to have it, I just forget the use case.
func (p *Participant) WalletIDs(_ struct{}, wl *[]state.WalletID) (err error) {
	p.engineLock.RLock()
	*wl = p.engine.WalletList()
	p.engineLock.RUnlock()
	return
}
