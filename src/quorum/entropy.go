package quorum

import (
	"fmt"
	"siacrypto"
)

const (
	EntropyVolume int = 32
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
	truncatedHash, err := siacrypto.CalculateTruncatedHash(q.seed[:])
	q.seed = Entropy(truncatedHash)
	return
}

// siblings are processed in a random order each block, determined by the
// entropy for the block. siblingOrdering() deterministically picks that
// order, using entropy from the state.
func (q *Quorum) SiblingOrdering() (siblingOrdering []byte) {
	q.lock.RLock()
	defer q.lock.RUnlock()

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
			// error - not sure what to do here
			continue
		}
		tmp := siblingOrdering[newIndex]
		siblingOrdering[newIndex] = siblingOrdering[i]
		siblingOrdering[i] = tmp
	}

	return
}

func (q *Quorum) IntegrateSiblingEntropy(e Entropy) {
	q.lock.Lock()
	defer q.lock.Unlock()

	th, err := siacrypto.CalculateTruncatedHash(append(q.germ[:], e[:]...))
	if err != nil {
		// hmm, error
		return
	}
	copy(q.seed[:], th[:])
}

func (q *Quorum) IntegrateGerm() {
	q.lock.Lock()
	defer q.lock.Unlock()

	copy(q.seed[:], q.germ[:])
}
