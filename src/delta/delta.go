// The delta layer manages inputs determined by the concensus layer and makes
// corresponding changes to the wallet layer.
package delta

import (
	"quorum"
	"sync"
)

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
