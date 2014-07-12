package quorum

import (
	"encoding/json"
	"siacrypto"
)

// Contains all of the general data for a quorum, and is an object that is
// specified to get sent over a wire. Anything that does not have an alternate
// encoding is put into this struct. Example of objects with alternate codings
// include wallets and events. This struct is meant to be small and to be sent
// over the wire as a complete entity, without needing to be broken up or
// buffered.
type QuorumMetadata struct {
	Siblings     [QuorumSize]*Sibling
	Seed         Entropy
	EventCounter uint32
	StoragePrice Balance
	Parent       siacrypto.Hash
	Height       uint32
}

func (q *Quorum) MarshalMetaData() (b []byte, err error) {
	b, err = json.Marshal(q.metaData)
	return
}
