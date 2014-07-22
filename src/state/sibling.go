package state

import (
	"network"
	"siacrypto"
)

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings.
type Sibling struct {
	Active    bool
	Index     byte
	Address   network.Address
	PublicKey siacrypto.PublicKey
	WalletID  WalletID
}

// Removes a sibling from the list of siblings
func (s *State) TossSibling(i byte) {
	s.Metadata.Siblings[i] = *new(Sibling)
}
