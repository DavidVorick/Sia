// The delta layer manages inputs determined by the concensus layer and makes
// corresponding changes to the wallet layer.
package delta

import (
	"state"
	"sync"
)

// The Engine struct has all the fields that enable basic operations at the
// delta level of the program. It's the 'master data structure' at this layer
// of abstraction.
//
// recentHistoryHead needs to be initialized to ^uint32(0)
// activeHistoryHead needs to be initialized to ^uint32(0) - uint32(SnapshotLenght - 1)
type Engine struct {
	// The State
	state state.State

	// Engine Variables
	filePrefix string

	// Snapshot Variables
	recentHistoryHead   uint32
	activeHistoryHead   uint32
	activeHistoryLength uint32
	historyLock         sync.RWMutex
}

func (e *Engine) SetFilePrefix(prefix string) {
	e.filePrefix = prefix
	walletPrefix := prefix + ".wallet"
	e.state.SetWalletPrefix(walletPrefix)
}

func (e *Engine) Metadata() state.StateMetadata {
	return e.state.Metadata
}

// NewBootstrapEngine() returns an engine that has its variables set so that
// the engine can function as the first sibling in a quorum. This requires a
// call to NewBootstrapState()
func (e *Engine) BootstrapEngine(sib *state.Sibling) (err error) {
	err = e.state.BootstrapState(sib)
	return
}
