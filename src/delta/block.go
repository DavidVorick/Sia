package delta

import (
	"quorum"
	"quorum/script"
	"siacrypto"
)

const (
	SnapshotLength         = 3 // number of blocks separating each snapshot
	BlockHistoryHeaderSize = 4 + SnapshotLen*4 + siacrypto.HashSize*SnapshotLen
)

// The heartbeat is the set of information that siblings are required to submit
// every block. Each block contains an array of [quorum.QuorumSize] heartbeats,
// and sets the value to 'nil' if nothing was submitted.
type Heartbeat struct {
	Entropy quorum.Entropy
	// storage proof
	Signature siacrypto.Signature
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
	Heartbeats [quorum.QuorumSize]*Heartbeat // using pointers enables setting Heartbeats to nil

	// Aggregate of non-required information submitted to the quorum
	ScriptInputs       []script.ScriptInput
	UploadAdvancements []quorum.UploadAdvancement
}
