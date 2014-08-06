package state

import (
	"siacrypto"
)

const (
	// EntropyVolume is the size of the Entropy type in byte.
	EntropyVolume int = 32
)

// Entropy is highly random data that is used to seed a PRNG.
type Entropy [EntropyVolume]byte

// MergeExternalEntropy takes as input some entropy (assumed to be the
// external source of entropy) and appends it to the Germ. The Germ then
// becomes the new Seed.
func (s *State) MergeExternalEntropy(e Entropy) {
	s.Metadata.Seed = Entropy(siacrypto.HashBytes(append(s.Metadata.Germ[:], e[:]...)))
}

// Use the entropy stored in the state to generate a random integer [low, high)
// randInt only runs during compile(), when the mutexes are already locked
/*func (s *State) randInt(low int, high int) (randInt int, err error) {
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
	hash := siacrypto.HashBytes(s.Metadata.Seed[:])
	s.Metadata.Seed = Entropy(hash)
	return
}*/
