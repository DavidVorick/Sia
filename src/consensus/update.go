package consensus

import (
	"delta"
)

// An Update is the set of information sent by each participant during
// consensus. This information includes the heartbeat, which contains required
// information for being a part of the quorum, and it contains optional
// information such as script inputs.
type Update struct {
	Heartbeat delta.Heartbeat

	// optional stuff
}
