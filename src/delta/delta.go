// The delta layer manages inputs determined by the concensus layer and makes
// corresponding changes to the wallet layer.
package delta

import (
	"state"
)

const (
	BootstrapWalletID = 0
)

// The Engine struct has all the fields that enable basic operations at the
// delta level of the program. It's the 'master data structure' at this layer
// of abstraction.
//
// SaveSnapshot() should be called upon initialzation
// recentHistoryHead needs to be initialized to ^uint32(0)
// avtiveHistoryLength should be initialized to SnapshotLength
// activeHistoryHead needs to be initialized to ^uint32(0) - (SnapshotLength-1), because the turnover will result in a new blockhistory file being created.
type Engine struct {
	// The State
	state state.State

	// Engine Variables
	filePrefix string

	// Snapshot Variables
	recentHistoryHead   uint32
	activeHistoryHead   uint32
	activeHistoryLength uint32
}

func (e *Engine) SetFilePrefix(prefix string) {
	e.filePrefix = prefix
	walletPrefix := prefix + ".wallet."
	e.state.SetWalletPrefix(walletPrefix)
}

func (e *Engine) Metadata() state.StateMetadata {
	return e.state.Metadata
}

func (e *Engine) Initialize(filePrefix string) {
	e.SetFilePrefix(filePrefix)
	return
}

// NewBootstrapEngine() returns an engine that has its variables set so that
// the engine can function as the first sibling in a quorum. This requires a
// call to NewBootstrapState()
func (e *Engine) Bootstrap(sib state.Sibling) (err error) {
	// Create the bootstrap wallet, which acts as a fountain to get the economy
	// started.
	w := state.Wallet{
		ID:      BootstrapWalletID,
		Balance: state.NewBalance(0, 25000000),
		Script:  BootstrapScript,
	}
	err = e.state.InsertWallet(w)
	if err != nil {
		return
	}

	// Create a walle with the default script for the sibling to use.
	defaultScript := DefaultScript(sib.PublicKey)
	sibWallet := state.Wallet{
		ID:      sib.WalletID,
		Balance: state.NewBalance(0, 1000000),
		Script:  defaultScript,
	}
	e.state.AddSibling(&sibWallet, sib)

	e.SaveSnapshot()
	e.recentHistoryHead = ^uint32(0)
	e.activeHistoryHead = ^uint32(0) - (SnapshotLength - 1)
	e.activeHistoryLength = SnapshotLength
	return
}
