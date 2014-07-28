package consensus

import (
	"delta"
	"state"
)

// Block is an RPC call that returns a block of a specific height. Participants
// only keep a history of so many blocks, so asking for future blocks or
// expired blocks will return an error.
func (p *Participant) Block(blockHeight uint32, block *delta.Block) (err error) {
	*block, err = p.engine.LoadBlock(blockHeight)
	return
}

// Metadata is an RPC call that returns the current state metadata, found in
// p.engine.state.Metadata
func (p *Participant) Metadata(_ struct{}, smd *state.StateMetadata) (err error) {
	*smd = p.engine.Metadata()
	return
}

// UpdateSegment is an RPC call that allows hosts to submit diffs that match
// uploads that have been confirmed by consensus.
func (p *Participant) UpdateSegment(sd delta.SegmentDiff, _ *struct{}) (err error) {
	err = p.engine.UpdateSegment(sd)
	return
}

// Not sure what the use is for this, mostly wallets are downloaded via
// snapshots. Doesn't hurt to have it, I just forget the use case.
func (p *Participant) WalletIDs(_ struct{}, wl *[]state.WalletID) (err error) {
	*wl = p.engine.WalletList()
	return
}
