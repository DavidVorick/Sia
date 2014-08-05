package consensus

import (
	"delta"
	"state"
)

// Block is an RPC that returns a block of a specific height. Participants
// only keep a history of so many blocks, so asking for future blocks or
// expired blocks will return an error.
func (p *Participant) Block(blockHeight uint32, block *delta.Block) (err error) {
	*block, err = p.engine.LoadBlock(blockHeight)
	return
}

// Metadata is an RPC that returns the current state metadata.
func (p *Participant) Metadata(_ struct{}, smd *state.StateMetadata) (err error) {
	*smd = p.engine.Metadata()
	return
}

// UpdateSegment is an RPC that allows hosts to submit diffs that match
// updates that have been confirmed by consensus.
func (p *Participant) UpdateSegment(sd delta.SegmentDiff, accepted *bool) (err error) {
	*accepted, err = p.engine.UpdateSegment(sd)

	if *accepted {
		// Submit a notification to the quorum that a match has been uploaded.
		newAdvancement := state.UpdateAdvancement{
			Index:    p.engine.SiblingIndex(),
			UpdateID: sd.UpdateID,
		}
		p.updateAdvancementsLock.Lock()
		p.updateAdvancements = append(p.updateAdvancements, newAdvancement)
		p.updateAdvancementsLock.Unlock()
	}

	return
}

// Not sure what the use is for this, mostly wallets are downloaded via
// snapshots. Doesn't hurt to have it, I just forget the use case.
func (p *Participant) WalletIDs(_ struct{}, wl *[]state.WalletID) (err error) {
	*wl = p.engine.WalletList()
	return
}
