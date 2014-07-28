package consensus

import (
	"state"
)

// Many of the synchronize RPC calls take a height as input, corresponding to
// the height of the snapshot being referenced.  Only certain heights have
// valid snapshots, picking an incorrect height will result in an error.

/*
type SynchronizeConsensus struct {
	CurrentStep byte
	// Heartbeats
}
*/

/*
// SynchronizeConsensus is an RPC call that returns all of the variables needed
// to be up-to-speed with the current round of consensus. This includes all of
// the heartbeats that have been received as well as the current step that the
// algorithm is on.
//
// This function is currently incomplete.
func (p *Participant) SynchronizeConsensus(_ struct{}, sc *SynchronizeConsensus) (err error) {
	sc.CurrentStep = p.currentStep
	return
}
*/

// SnapshotMetadata is an RPC call that returns the metadata of the snapshot
// associated with the input height.
func (p *Participant) SnapshotMetadata(snapshotHead uint32, snapshotMetadata *state.StateMetadata) (err error) {
	*snapshotMetadata, err = p.engine.LoadSnapshotMetadata(snapshotHead)
	return
}

// SnapshotWalletList is an RPC call that returns a list of the WalletIDs of
// every single wallet contained within the snapshot.
func (p *Participant) SnapshotWalletList(snapshotHead uint32, walletList *[]state.WalletID) (err error) {
	*walletList, err = p.engine.LoadSnapshotWalletList(snapshotHead)
	return
}

//  SnapshotWalletInput asks for a particular wallet from a particular
//  snapshot. RPC only supports a single variable as input, so a separate
//  struct is needed to support both variables in the RPC call.
type SnapshotWalletInput struct {
	SnapshotHead uint32
	WalletID     state.WalletID
}

func (p *Participant) SnapshotWallet(swi SnapshotWalletInput, wallet *state.Wallet) (err error) {
	*wallet, err = p.engine.LoadSnapshotWallet(swi.SnapshotHead, swi.WalletID)
	return
}
