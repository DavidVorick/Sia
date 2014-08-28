package delta

import (
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

const (
	// FountainWalletID is a wallet that generates miniscule volumes of
	// coins through which people can join the network for free.
	FountainWalletID = 0
)

// Bootstrap returns an engine that has its variables set so that
// the engine can function as the first sibling in a quorum.
func (e *Engine) Bootstrap(sib state.Sibling, tetherWalletPublicKey siacrypto.PublicKey) (err error) {
	// Create the bootstrap wallet, which acts as a fountain to get the economy
	// started.
	err = e.state.InsertWallet(state.Wallet{
		ID:      FountainWalletID,
		Balance: state.NewBalance(25000000),
		Script:  FountainScript,
	}, true)
	if err != nil {
		return
	}

	// Create a wallet with the default script for the sibling to use.
	sibWallet := state.Wallet{
		ID:      sib.WalletID,
		Balance: state.NewBalance(1000000),
		Script:  DefaultScript(tetherWalletPublicKey),
	}
	err = e.state.InsertWallet(sibWallet, true)
	if err != nil {
		return
	}
	e.AddSibling(&sibWallet, sib)
	e.state.Metadata.Siblings[0].Status = 0

	e.recentHistoryHead = ^uint32(0)
	e.state.Metadata.RecentSnapshot = ^uint32(0) - (SnapshotLength - 1)
	e.activeHistoryLength = SnapshotLength
	return
}

// BootstrapJoinSetup currently just sets the snapshot variables when a
// participant is just joining the network.
func (e *Engine) BootstrapJoinSetup() (err error) {
	// Set e.activeHistoryLength to SnapshotLength. It's a bit of a hack,
	// but it signals to the snapshot code that a new blockhistory file
	// needs to be created.
	e.activeHistoryLength = SnapshotLength
	return
}

// BootstrapSetMetadata is functionally equivalent to 'SetMetadata', but it's a
// function that should _only_ be called during the bootstrapping process,
// which is why it has the extra word in the name. I wanted to implement
// something like an initialize that would take a bunch of wallets and a
// metadata as input, but that could take a massive amount of memory. I wasn't
// certain about the best way to approach the problem, so this is the solution
// I've picked for the time being.
func (e *Engine) BootstrapSetMetadata(md state.Metadata) {
	e.state.Metadata = md
}

// BootstrapInsertWallet is functionally equivalent to 'InsertWallet', except
// that this function should _only_ be called during bootstrapping. See
// 'BootstrapSetMetadata' comment for more details.
func (e *Engine) BootstrapInsertWallet(w state.Wallet) (err error) {
	err = e.state.InsertWallet(w, false)
	return
}
