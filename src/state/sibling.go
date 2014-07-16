package state

import (
	"network"
	"siacrypto"
)

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings.
type Sibling struct {
	Index     byte
	Address   network.Address
	PublicKey *siacrypto.PublicKey
	WalletID  WalletID
}

// Removes a sibling from the list of siblings
func (s *State) TossSibling(i byte) {
	s.Metadata.Siblings[i] = nil
}

// Sibling returns true if the address and publicKey fields are identical
func (s0 *Sibling) Compare(s1 *Sibling) bool {
	// false if either sibling is nil
	if s0 == nil || s1 == nil {
		return false
	}

	// return false if the addresses are not equal
	if s0.Address != s1.Address {
		return false
	}

	// return false if the public keys are not equivalent
	compare := s0.PublicKey.Compare(s1.PublicKey)
	if compare != true {
		return false
	}

	return true
}
