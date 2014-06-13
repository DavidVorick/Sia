package quorum

import (
	"fmt"
	"siacrypto"
)

const (
	EntropyVolume int = 32 // in bytes
)

type Entropy [EntropyVolume]byte

// Use the entropy stored in the state to generate a random integer [low, high)
// randInt only runs during compile(), when the mutexes are already locked
func (q *Quorum) randInt(low int, high int) (randInt int, err error) {
	q.lock.Lock()
	defer q.lock.Unlock()

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
	truncatedHash := siacrypto.CalculateHash(q.seed[:])
	q.seed = Entropy(truncatedHash)
	return
}

// siblings are processed in a random order each block, determined by the
// entropy for the block. siblingOrdering() deterministically picks that
// order, using entropy from the state.
func (q *Quorum) SiblingOrdering() (siblingOrdering []byte) {
	// create an in-order list of siblings, leaving out nil siblings
	q.lock.RLock()
	for i, s := range q.siblings {
		if s != nil {
			siblingOrdering = append(siblingOrdering, byte(i))
		}
	}
	q.lock.RUnlock()

	// shuffle the list of siblings
	for i := range siblingOrdering {
		newIndex, err := q.randInt(i, len(siblingOrdering))
		if err != nil {
			// error
			continue
		}
		tmp := siblingOrdering[newIndex]
		siblingOrdering[newIndex] = siblingOrdering[i]
		siblingOrdering[i] = tmp
	}

	return
}

// It was a tough decision to move this functionality from the participant to
// the quorum. But really I don't think that a participant should have access
// to the field that will eventually become entropy.
func (q *Quorum) IntegrateSiblingEntropy(e Entropy) {
	q.lock.Lock()
	defer q.lock.Unlock()

	th := siacrypto.CalculateHash(append(q.germ[:], e[:]...))
	copy(q.germ[:], th[:])
}

// This will eventually be replaced with functionality that considers external
// entropy such as Aiza.
func (q *Quorum) IntegrateGerm() {
	q.lock.Lock()
	defer q.lock.Unlock()

	copy(q.seed[:], q.germ[:])
}
