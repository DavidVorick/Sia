package consensus

import (
	"delta"
	"siacrypto"
)

// An Update is the set of information sent by each participant during
// consensus. This information includes the heartbeat, which contains required
// information for being a part of the quorum, and it contains optional
// information such as script inputs.
type Update struct {
	Heartbeat          delta.Heartbeat
	HeartbeatSignature siacrypto.Signature

	// optional stuff
}
