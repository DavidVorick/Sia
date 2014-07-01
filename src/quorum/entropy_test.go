package quorum

import (
	"testing"
)

func TestRandInt(t *testing.T) {
	q := new(Quorum)

	// call randInt under normal conditions
	previousSeed := q.seed
	randInt, err := q.randInt(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if randInt < 0 || randInt >= 5 {
		t.Fatal("randInt returned but is not between the bounds")
	}

	// check that s.CurrentEntropy flipped to next value
	if previousSeed == q.seed {
		t.Error(previousSeed)
		t.Error(q.seed)
		t.Fatal("When calling randInt, s.CurrentEntropy was not changed")
	}

	// trigger an error condition
	randInt, err = q.randInt(0, 0)
	if err == nil {
		t.Fatal("Randint(0,0) should return a bounds error")
	}

	// fuzzing, skip for short tests
	if testing.Short() {
		t.Skip()
	}

	low := 0
	high := int(QuorumSize)
	for i := 0; i < 100000; i++ {
		randInt, err = q.randInt(low, high)
		if err != nil {
			t.Fatal("randInt fuzzing error: ", err, " low: ", low, " high: ", high)
		}

		if randInt < low || randInt >= high {
			t.Fatal("randInt fuzzing: ", randInt, " produced, expected number between ", low, " and ", high)
		}
	}
}

func TestSiblingOrdering(t *testing.T) {
	q := new(Quorum)

	// add QuorumSize siblings to the Quorum, each time calling SiblingOrdering
	// and verifying that the correct number of siblings are present
	for i := byte(0); i < QuorumSize; i++ {
		sibling := Sibling{
			index: i,
		}

		q.siblings[i] = &sibling

		siblings := q.SiblingOrdering()
		if len(siblings) != int(i+1) {
			t.Error("SiblingOrdering producing wrong number of siblings")
		}
	}
}

func TestGermAndSeed(t *testing.T) {
	q := new(Quorum)

	// Integrate a seed into the germ and then integrate the germ, see that each
	// member variable of the quorum updates accordingly.
	oldGerm := q.germ
	var e Entropy
	q.IntegrateSiblingEntropy(e)
	if oldGerm == q.germ {
		t.Error("Germ did not update when integrating sibling entropy")
	}
	oldSeed := q.seed
	q.IntegrateGerm()
	if oldSeed == q.seed {
		t.Error("Seed did not update when integrating germ")
	}
}
