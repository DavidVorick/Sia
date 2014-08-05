// Package delta manages inputs determined by the consensus layer and makes
// corresponding changes to the state layer.
package delta

// I don't like many aspects of this file. It's a whole bunch of getters and
// setters, which I feel largely implies a poor design. I would, in general,
// consider this file to be the 'junk' file of the delta package, and for the
// most part I think it's the worst aspect of the reorganization. I hate to say
// it, but I feel like Sia could use further engineering. It's a complex system
// and its current designs are passable, but not elegant, clean, or beautiful.

import (
	"state"
)

const (
	// BootstrapWalletID is the wallet ID used by the siacoin 'fountain'.
	BootstrapWalletID = 0
)

// The Engine struct has all the fields that enable basic operations at the
// delta level of the program. It's the 'master data structure' at this layer
// of abstraction.
//
// SaveSnapshot() should be called upon initialzation
// recentHistoryHead needs to be initialized to ^uint32(0)
// activeHistoryLength should be initialized to SnapshotLength
// activeHistoryHead needs to be initialized to ^uint32(0) - (SnapshotLength-1), because the turnover will result in a new blockhistory file being created.
type Engine struct {
	// The State
	state state.State

	// Engine Variables
	filePrefix   string
	siblingIndex byte

	// Upload Variables
	completedUpdates map[state.UpdateID]bool

	// Snapshot Variables
	recentHistoryHead   uint32
	activeHistoryHead   uint32
	activeHistoryLength uint32
}

// SetFilePrefix is a setter for the Engine.filePrefix field.
// It also sets the walletPrefix field of the state object.
func (e *Engine) SetFilePrefix(prefix string) {
	e.filePrefix = prefix
	walletPrefix := prefix + ".wallet."
	e.state.SetWalletPrefix(walletPrefix)
}

// Metadata is a getter that returns the state.Metadata object.
func (e *Engine) Metadata() state.StateMetadata {
	return e.state.Metadata
}

// SiblingIndex is a getter that returns the engine's sibling index.
func (e *Engine) SiblingIndex() byte {
	return e.siblingIndex
}

// WalletList is a pass-along function so that the wallet list of the state can be accessed
// by instances containing the engine.
func (e *Engine) WalletList() []state.WalletID {
	return e.state.WalletList()
}

// Initialize sets various fields of the Engine object.
func (e *Engine) Initialize(filePrefix string, siblingIndex byte) {
	e.SetFilePrefix(filePrefix)
	e.siblingIndex = siblingIndex
	return
}

// Bootstrap returns an engine that has its variables set so that
// the engine can function as the first sibling in a quorum.
func (e *Engine) Bootstrap(sib state.Sibling) (err error) {
	// Create the bootstrap wallet, which acts as a fountain to get the economy started.
	err = e.state.InsertWallet(state.Wallet{
		ID:      BootstrapWalletID,
		Balance: state.NewBalance(0, 25000000),
		Script:  BootstrapScript,
	})
	if err != nil {
		return
	}

	// Create a walle with the default script for the sibling to use.
	sibWallet := state.Wallet{
		ID:      sib.WalletID,
		Balance: state.NewBalance(0, 1000000),
		Script:  DefaultScript(sib.PublicKey),
	}
	err = e.state.InsertWallet(sibWallet)
	if err != nil {
		return
	}
	e.AddSibling(&sibWallet, sib)

	e.saveSnapshot()
	e.recentHistoryHead = ^uint32(0)
	e.activeHistoryHead = ^uint32(0) - (SnapshotLength - 1)
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
func (e *Engine) BootstrapSetMetadata(md state.StateMetadata) {
	e.state.Metadata = md
}

// BootstrapInsertWallet is functionally equivalent to 'InsertWallet', except
// that this function should _only_ be called during bootstrapping. See
// 'BootstrapSetMetadata' comment for more details.
func (e *Engine) BootstrapInsertWallet(w state.Wallet) (err error) {
	err = e.state.InsertWallet(w)
	return
}
