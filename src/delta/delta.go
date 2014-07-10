// The delta layer manages inputs determined by the concensus layer and makes
// corresponding changes to the wallet layer.
package delta

import (
	"quorum"
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
	quorum quorum.Quorum

	// Engine Variables
	filePrefix string

	// Snapshot Variables
	recentHistoryHead   uint32
	activeHistoryHead   uint32
	activeHistoryLength uint32
	historyLock         sync.RWMutex
}
