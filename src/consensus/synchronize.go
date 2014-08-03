package consensus

import (
	"delta"
	"state"
)

// A SynchronizeConsensusResp is a simple struct used in the SynchronizeConsensus RPC.
type SynchronizeConsensusResp struct {
	CurrentStep byte
	// Heartbeats
}

// SynchronizeConsensus returns all of the variables needed to be up-to-speed
// with the current round of consensus. This includes all of the heartbeats
// that have been received as well as the current step that the algorithm is
// on.
func (p *Participant) SynchronizeConsensus(_ struct{}, scr *SynchronizeConsensusResp) (err error) {
	scr.CurrentStep = p.currentStep
	return
}

// Metadata is an RPC that returns the engine's StateMetadata object.
func (p *Participant) Metadata(_ struct{}, smd *state.StateMetadata) (err error) {
	*smd = p.engine.Metadata()
	return
}

// WalletIDs is an RPC that returns the engine's slice of WalletIDs.
func (p *Participant) WalletIDs(_ struct{}, wl *[]state.WalletID) (err error) {
	*wl = p.engine.WalletList()
	return
}

// SnapshotMetadata is an RPC that returns the engine's StateMetadata object corresponding to a given snapshot head.
func (p *Participant) SnapshotMetadata(snapshotHead uint32, snapshotMetadata *state.StateMetadata) (err error) {
	*snapshotMetadata, err = p.engine.LoadSnapshotMetadata(snapshotHead)
	return
}

// SnapshotWalletList is an RPC that returns the engine's slice of WalletIDs corresponding to a given snapshot head.
func (p *Participant) SnapshotWalletList(snapshotHead uint32, walletList *[]state.WalletID) (err error) {
	*walletList, err = p.engine.LoadSnapshotWalletList(snapshotHead)
	return
}

// A SnapshotWalletArg is a simple struct used in the SnapshotWallet RPC.
type SnapshotWalletArg struct {
	SnapshotHead uint32
	WalletID     state.WalletID
}

// SnapshotWallet is an RPC that returns the Wallet corresponding to a given snapshot head and WalletID.
func (p *Participant) SnapshotWallet(swa SnapshotWalletArg, wallet *state.Wallet) (err error) {
	*wallet, err = p.engine.LoadSnapshotWallet(swa.SnapshotHead, swa.WalletID)
	return
}

// Block is an RPC that returns the Block corresponding to a given block height.
func (p *Participant) Block(blockHeight uint32, block *delta.Block) (err error) {
	*block, err = p.engine.LoadBlock(blockHeight)
	return
}
