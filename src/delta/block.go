package delta

import (
	"siacrypto"
	"state"
)

const (
	SnapshotLength = 3 // blocks separating each snapshot (probably needs a different name)
)

// The heartbeat is the set of information that siblings are required to submit
// every block. Each block contains an array of [state.QuorumSize] heartbeats,
// and sets the value to 'nil' if nothing was submitted.
type Heartbeat struct {
	ParentBlock siacrypto.Hash
	Entropy     state.Entropy
	// storage proof
}

// A block contains all the data that is necessary to move the quorum from one
// state to the next. It contains a height and a parent block, as well as a
// parent quorum. These values enable the quorum to verify that the block is
// consistent with the current quorum and is not a block that is targeted
// toward a fork.
type Block struct {
	// Meta data for the block
	Height      uint32
	ParentBlock siacrypto.Hash
	// parentQuorum

	// Heartbeats for each sibling
	Heartbeats          [state.QuorumSize]Heartbeat
	HeartbeatSignatures [state.QuorumSize]siacrypto.Signature

	// Aggregate of non-required information submitted to the quorum
	ScriptInputs          []ScriptInput
	UpdateAdvancements    []state.UpdateAdvancement
	AdvancementSignatures []siacrypto.Signature
}
