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
func (s *Sibling) PublicKey() *siacrypto.PublicKey {
	return s.publicKey
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

// Sibling returns true if the address and publicKey fields are identical
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

// EncodedSiblings returns a []byte that contains sufficient information to
// decode a [QuorumSize]*Sibling, which is necessary because gob can't encode
// an array of siblings. This function doesn't act on a quorum because external
// packages need to be able to receive siblings without having an entire quorum
// attatched. I'm not sure if it's the best design decision.
func EncodeSiblings(siblings [QuorumSize]*Sibling) (encodedSiblings []byte, err error) {
	var siblingSlice []*Sibling
	for i := range siblings {
		if siblings[i] != nil {
			siblingSlice = append(siblingSlice, siblings[i])
		}
	}

	b := new(bytes.Buffer)
	encoder := gob.NewEncoder(b)
	err = encoder.Encode(siblingSlice)
	if err != nil {
		return
	}
	encodedSiblings = b.Bytes()
	return
}

func DecodeSiblings(encodedSiblings []byte) (siblings [QuorumSize]*Sibling, err error) {
	b := bytes.NewBuffer(encodedSiblings)
	decoder := gob.NewDecoder(b)

	var siblingSlice []*Sibling
	err = decoder.Decode(&siblingSlice)
	if err != nil {
		return
	}
	for i := range siblingSlice {
		siblings[siblingSlice[i].index] = siblingSlice[i]
	}
	return
}
