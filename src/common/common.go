// Common contains structs, data, and interfaces that
// need to be referenced by other packages
package common

import (
	"time"
)

const (
	// How many bytes of entropy must be produced each entropy cycle
	EntropyVolume int = 32

	// How big a single segment of data is for a host, in bytes
	MinSegmentSize int = 512
	MaxSegmentSize int = 1048576 // 1 MB

	// How many participants are in each quorum
	// This number is chosen to minimize the probability of a single quorum
	// 	becoming more than 80% compromised by an attacker controlling 1/2 of
	// 	the network.
	QuorumSize int = 4

	// How long a single step in the consensus algorithm takes
	StepDuration time.Duration = 3 * time.Second
)

type Entropy [EntropyVolume]byte
