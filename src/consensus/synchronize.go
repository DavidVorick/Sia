package consensus

import (
	"state"
)

type SynchronizeConsensus struct {
	CurrentStep byte
	// Heartbeats
}

// SynchronizeConsensus returns all of the variables needed to be up-to-speed
// with the current round of consensus. This includes all of the heartbeats
// that have been received as well as the current step that the algorithm is
// on.
func (p *Participant) SynchronizeConsensus(_ struct{}, sc *SynchronizeConsensus) (err error) {
	sc.CurrentStep = p.currentStep
	return
}

func (p *Participant) Metadata(_ struct{}, smd *state.StateMetadata) (err error) {
	*smd = p.engine.Metadata()
	return
}

func (p *Participant) WalletIDs(_ struct{}, wl *[]state.WalletID) (err error) {
	*wl = p.engine.WalletList()
	return
}

func (p *Participant) RecentSnapshot(_ struct{}, height *uint32) (err error) {
	*height = p.engine.ActiveHistoryHead()
	return
}

func (p *Participant) SnapshotMetadata(snapshotHead uint32, snapshotMetadata *state.StateMetadata) (err error) {
	*snapshotMetadata, err = p.engine.LoadSnapshotMetadata(snapshotHead)
	return
}

func (p *Participant) SnapshotWalletList(snapshotHead uint32, walletList *[]state.WalletID) (err error) {
	*walletList, err = p.engine.LoadSnapshotWalletList(snapshotHead)
	return
}
