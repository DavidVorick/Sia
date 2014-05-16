package quorum

import (
	"bytes"
	"common"
	"common/crypto"
	"common/log"
	"crypto/ecdsa"
	"encoding/gob"
	"fmt"
	"sync"
)

// Identifies other members of the quorum
type Sibling struct {
	index     byte // not sure that this is the appropriate place for this variable
	address   common.Address
	publicKey *crypto.PublicKey
}

type quorum struct {
	// Network Variables
	siblings     [common.QuorumSize]*Sibling // list of all siblings in quorum
	siblingsLock sync.RWMutex                // prevents race conditions
	// meta quorum

	// Heartbeat Variables
	heartbeats     [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat
	heartbeatsLock sync.Mutex
	// file proofs stage 2

	// Compile Variables
	currentEntropy  common.Entropy // Used to generate random numbers during compilation
	upcomingEntropy common.Entropy // Used to compute entropy for next block

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

// Returns true if the values of the siblings are equivalent
func (s0 *Sibling) compare(s1 *Sibling) bool {
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

// siblings are processed in a random order each block, determined by the
// entropy for the block. siblingOrdering() deterministically picks that
// order, using entropy from the state.
func (q *quorum) siblingOrdering() (siblingOrdering [common.QuorumSize]byte) {
	// create an in-order list of siblings
	for i := range siblingOrdering {
		siblingOrdering[i] = byte(i)
	}

	// shuffle the list of siblings
	for i := range siblingOrdering {
		newIndex, err := q.randInt(i, common.QuorumSize)
		if err != nil {
			log.Fatalln(err)
		}
		tmp := siblingOrdering[newIndex]
		siblingOrdering[newIndex] = siblingOrdering[i]
		siblingOrdering[i] = tmp
	}

	return
}

// Removes all traces of a sibling from the State
func (q *quorum) tossSibling(pi byte) {
	// remove from s.Siblings
	q.siblings[pi] = nil

	// nil map in s.Heartbeats
	q.heartbeats[pi] = nil
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
	epk := (*ecdsa.PublicKey)(s.publicKey)
	if epk.X == nil {
		err = fmt.Errorf("Cannot encode nil value s.publicKey.X")
		return
	}
	if epk.Y == nil {
		err = fmt.Errorf("Cannot encode nil value s.publicKey.Y")
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
	gobSibling = w.Bytes()
	return
}

func (s *Sibling) GobDecode(gobSibling []byte) (err error) {
	if s == nil {
		err = fmt.Errorf("Cannot decode into nil sibling")
		return
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
	return
}

// Update the state according to the information presented in the heartbeat
// processHeartbeat uses return codes for testing purposes
func (q *quorum) processHeartbeat(hb *heartbeat, i byte) (err error) {
	print("Confirming Sibling")
	println(i)

	// Add the entropy to UpcomingEntropy
	th, err := crypto.CalculateTruncatedHash(append(q.upcomingEntropy[:], hb.entropy[:]...))
	q.upcomingEntropy = common.Entropy(th)

	return
}

// Use the entropy stored in the state to generate a random integer [low, high)
// randInt only runs during compile(), when the mutexes are already locked
func (q *quorum) randInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 8
		rollingInt += int(q.currentEntropy[i])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(q.currentEntropy[:])
	q.currentEntropy = common.Entropy(truncatedHash)
	return
}
