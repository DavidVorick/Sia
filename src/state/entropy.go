package state

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
func (s *State) randInt(low int, high int) (randInt int, err error) {
	// verify there's a gap between the numbers
	if low == high {
		err = fmt.Errorf("low and high cannot be the same number")
		return
	}

	// Convert CurrentEntropy into an int
	rollingInt := 0
	for i := 0; i < 4; i++ {
		rollingInt = rollingInt << 8
		rollingInt += int(s.Metadata.Seed[i])
	}

	randInt = (rollingInt % (high - low)) + low

	// Convert random number seed to next value
	hash := siacrypto.CalculateHash(s.Metadata.Seed[:])
	s.Metadata.Seed = Entropy(hash)
	return
}

// siblings are processed in a random order each block, determined by the
// entropy for the block. siblingOrdering() deterministically picks that
// order, using entropy from the state.
func (s *State) SiblingOrdering() (siblingOrdering []byte) {
	// create an in-order list of siblings, leaving out nil siblings
	for i, s := range s.Metadata.Siblings {
		if s != nil {
			siblingOrdering = append(siblingOrdering, byte(i))
		}
	}

	// shuffle the list of siblings
	for i := range siblingOrdering {
		newIndex, err := s.randInt(i, len(siblingOrdering))
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

/*
// It was a tough decision to move this functionality from the participant to
// the quorum. But really I don't think that a participant should have access
// to the field that will eventually become entropy.
func (s *State) IntegrateSiblingEntropy(e Entropy) {
	th := siacrypto.CalculateHash(append(s.germ[:], e[:]...))
	copy(s.germ[:], th[:])
}

// This will eventually be replaced with functionality that considers external
// entropy such as Aiza.
func (s *State) IntegrateGerm() {
	copy(s.seed[:], s.germ[:])
}*/
