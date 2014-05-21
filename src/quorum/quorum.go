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

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings.
type Sibling struct {
	index     byte
	address   common.Address
	publicKey *crypto.PublicKey
}

// A quorum is a set of data that is identical across all participants in the
// quorum. Data in the quorum can only be updated during a block, and the
// update must be deterministic and reversable.
type quorum struct {
	// Network Variables
	siblings     [common.QuorumSize]*Sibling
	siblingsLock sync.RWMutex
	// meta quorum

	// file proofs stage 1

	// Compile Variables
	seed common.Entropy // Used to generate random numbers during compilation

	// Batch management
	parent *batchNode
}

// Sibling.compare returns true if the values of each Sibling are equivalent
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
	return
}

// siblings are processed in a random order each block, determined by the
// entropy for the block. siblingOrdering() deterministically picks that
// order, using entropy from the state.
func (q *quorum) siblingOrdering() (siblingOrdering []byte) {
	// create an in-order list of siblings
	for i, s := range q.siblings {
		if s != nil {
			siblingOrdering = append(siblingOrdering, byte(i))
		}
	}

	// shuffle the list of siblings
	for i := range siblingOrdering {
		newIndex, err := q.randInt(i, len(siblingOrdering))
		if err != nil {
			log.Fatalln(err)
		}
		tmp := siblingOrdering[newIndex]
		siblingOrdering[newIndex] = siblingOrdering[i]
		siblingOrdering[i] = tmp
	}

	return
}

// Removes a sibling from the list of siblings
func (q *quorum) tossSibling(pi byte) {
	println("tossing sibling", pi)
	q.siblings[pi] = nil
}

// Update the state according to the information presented in the heartbeat
func (q *quorum) processHeartbeat(hb *heartbeat, seed common.Entropy) (newSiblings []*Sibling, newSeed common.Entropy, err error) {
	// add hopefuls to any available slots
	// quorum.siblings has already been locked by compile()
	j := 0
	for _, s := range hb.hopefuls {
		for j < common.QuorumSize {
			if q.siblings[j] == nil {
				println("placed hopeful at index", j)
				s.index = byte(j)
				q.siblings[s.index] = s
				newSiblings = append(newSiblings, s)
				break
			}
			j++
		}
	}

	// Add the entropy to newSeed
	th, err := crypto.CalculateTruncatedHash(append(seed[:], hb.entropy[:]...))
	newSeed = common.Entropy(th)

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
		rollingInt += int(q.seed[i])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	truncatedHash, err := crypto.CalculateTruncatedHash(q.seed[:])
	q.seed = common.Entropy(truncatedHash)
	return
}

// q.Status() enumerates the variables of the quorum in a human-readable output
func (q *quorum) Status() (b string) {
	b = "\tSiblings:\n"
	for _, s := range q.siblings {
		if s != nil {
			b += fmt.Sprintf("\t\t%v %v\n", s.index, s.address)
		}
	}

	b += fmt.Sprintf("\tSeed: %x\n", q.seed)
	return
}

// Only the siblings and entropy are encoded.
func (q *quorum) GobEncode() (gobQuorum []byte, err error) {
	// if q == nil, encode a zero quorum
	if q == nil {
		q = new(quorum)
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode network variabes
	// Only encode non-nil siblings
	var encSiblings []*Sibling
	for _, s := range q.siblings {
		if s != nil {
			encSiblings = append(encSiblings, s)
		}
	}
	err = encoder.Encode(encSiblings)
	if err != nil {
		return
	}

	// Encode compile variables
	err = encoder.Encode(q.seed)
	if err != nil {
		return
	}

	// Encode batch variables
	// tbi

	gobQuorum = w.Bytes()
	return
}

// Only the siblings and entropy are decoded.
func (q *quorum) GobDecode(gobQuorum []byte) (err error) {
	// if q == nil, make a new quorum and decode into that
	if q == nil {
		q = new(quorum)
	}

	r := bytes.NewBuffer(gobQuorum)
	decoder := gob.NewDecoder(r)

	// decode slice of siblings into the sibling array
	var encSiblings []*Sibling
	err = decoder.Decode(&encSiblings)
	if err != nil {
		return
	}
	for _, s := range encSiblings {
		q.siblings[s.index] = s
	}

	// decode compile variables
	err = decoder.Decode(&q.seed)
	if err != nil {
		return
	}

	// decode batch variables
	// tbi

	return
}
