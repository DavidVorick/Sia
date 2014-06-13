package quorum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"network"
	"siacrypto"
)

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings.
type Sibling struct {
	index     byte
	address   network.Address
	publicKey *siacrypto.PublicKey
	wallet    WalletID
}

// Getters for the private variables
func (s *Sibling) Index() byte {
	return s.index
}
func (s *Sibling) Address() network.Address {
	return s.address
}
func (s *Sibling) PublicKey() siacrypto.PublicKey {
	return *s.publicKey
}

// Sibling variables are kept private because they should not be changing
// unless the quorum is making changes to its structure and all siblings have
// access to the same set of changes.
func NewSibling(address network.Address, key *siacrypto.PublicKey) *Sibling {
	return &Sibling{
		index:     255,
		address:   address,
		publicKey: key,
	}
}

// Removes a sibling from the list of siblings
func (q *Quorum) TossSibling(i byte) {
	q.siblings[i] = nil
}

// Sibling.compare returns true if the values of each Sibling are equivalent
// The index field is not considered.
func (s0 *Sibling) Compare(s1 *Sibling) bool {
	// false if either sibling is nil
	if s0 == nil || s1 == nil {
		return false
	}

	// return false if the addresses are not equal
	if s0.address != s1.address {
		return false
	}

	// return false if the public keys are not equivalent
	compare := s0.publicKey.Compare(s1.publicKey)
	if compare != true {
		return false
	}

	return true
}

func (s *Sibling) GobEncode() (gobSibling []byte, err error) {
	// Error checking for nil values
	// Because public keys cannot be nil and are not valid as zero-values, a nil
	// participant cannot be encoded
	if s == nil {
		err = fmt.Errorf("Cannot encode nil sibling")
		return
	}
	if s.publicKey == nil {
		err = fmt.Errorf("Cannot encode nil value s.publicKey")
		return
	}

	// Encoding the sibling
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(s.index)
	if err != nil {
		return
	}
	err = encoder.Encode(s.address)
	if err != nil {
		return
	}
	err = encoder.Encode(s.publicKey)
	if err != nil {
		return
	}
	err = encoder.Encode(s.wallet)
	if err != nil {
		return
	}

	gobSibling = w.Bytes()
	return
}

func (s *Sibling) GobDecode(gobSibling []byte) (err error) {
	// if nil, make a new sibling object
	if s == nil {
		s = new(Sibling)
	}

	r := bytes.NewBuffer(gobSibling)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&s.index)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.address)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.publicKey)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.wallet)
	if err != nil {
		return
	}

	return
}
