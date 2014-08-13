package consensus

import (
	"github.com/NebulousLabs/Sia/state"
)

// Many of the synchronize RPC calls take a height as input, corresponding to
// the height of the snapshot being referenced.  Only certain heights have
// valid snapshots, picking an incorrect height will result in an error.

// SnapshotMetadata is an RPC that returns the engine's Metadata object
// corresponding to a given snapshot head.
func (p *Participant) SnapshotMetadata(snapshotHead uint32, snapshotMetadata *state.Metadata) (err error) {
	*snapshotMetadata, err = p.engine.LoadSnapshotMetadata(snapshotHead)
	return
}

// SnapshotWalletList is an RPC that returns the list of WalletIDs
// corresponding to a given snapshot head.
func (p *Participant) SnapshotWalletList(snapshotHead uint32, walletList *[]state.WalletID) (err error) {
	*walletList, err = p.engine.LoadSnapshotWalletList(snapshotHead)
	return
}

// A SnapshotWalletArg is a simple struct used in the SnapshotWallet RPC.
type SnapshotWalletArg struct {
	SnapshotHead uint32
	WalletID     state.WalletID
}

// SnapshotWallet is an RPC that returns the Wallet corresponding to a given
// snapshot head and WalletID.
func (p *Participant) SnapshotWallet(swa SnapshotWalletArg, wallet *state.Wallet) (err error) {
	*wallet, err = p.engine.LoadSnapshotWallet(swa.SnapshotHead, swa.WalletID)
	return
}
