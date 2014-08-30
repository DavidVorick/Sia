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
	"github.com/NebulousLabs/Sia/state"
)

// The Engine struct has all the fields that enable basic operations at the
// delta level of the program. It's the 'master data structure' at this layer
// of abstraction.
//
// - recentHistoryHead needs to be initialized to ^uint32(0).
// - activeHistoryLength should be initialized to SnapshotLength.
// - e.state.Metadata.RecentSnapshot needs to be initialized to ^uint32(0) - (SnapshotLength-1),
//   because the turnover will result in a new blockhistory file being created.
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
	activeHistoryLength uint32
}

// SetFilePrefix is a setter for the Engine.filePrefix field.
// It also sets the walletPrefix field of the state object.
func (e *Engine) SetFilePrefix(prefix string) {
	e.filePrefix = prefix
	walletPrefix := prefix + "wallet"
	e.state.SetWalletPrefix(walletPrefix)
}

func (e *Engine) SiblingIndex() byte {
	return e.siblingIndex
}

func (e *Engine) SetSiblingIndex(index byte) {
	// Other things might go here eventually.
	e.siblingIndex = index
}

func (e *Engine) Initialize(filePrefix string) {
	e.SetFilePrefix(filePrefix)
	e.state.Initialize()
}

// Metadata is a getter that returns the state.Metadata object.
func (e *Engine) Metadata() state.Metadata {
	return e.state.Metadata
}

func (e *Engine) Wallet(id state.WalletID) (w state.Wallet, err error) {
	w, err = e.state.LoadWallet(id)
	return
}

// WalletList is a pass-along function so that the wallet list of the state can
// be accessed by instances containing the engine.
func (e *Engine) WalletList() []state.WalletID {
	return e.state.WalletList()
}
