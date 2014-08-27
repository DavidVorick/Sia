package state

import (
	"github.com/NebulousLabs/Sia/siacrypto"
)

// Metadata contains all of the general data for a quorum, and is an object
// that is specified to get sent over a wire. Anything that does not have an
// alternate encoding is put into this struct. Examples of objects with
// alternate encodings include wallets and events. This struct is meant to be
// small and to be sent over the wire as a complete entity, without needing
// to be broken up or buffered.
type Metadata struct {
	Siblings       [QuorumSize]Sibling
	Germ           Entropy
	Seed           Entropy
	EventCounter   uint32
	StoragePrice   Balance
	ParentBlock    siacrypto.Hash
	Height         uint32
	RecentSnapshot uint32
	PoStorageSeed  siacrypto.Hash
}
