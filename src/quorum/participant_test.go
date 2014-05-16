package quorum

import (
	"common"
	"testing"
)

// Create a state, check the defaults
func TestCreateParticipant(t *testing.T) {
	// make sure CreateState does not cause errors
	p, err := CreateParticipant(common.NewZeroNetwork())
	if err != nil {
		t.Fatal(err)
	}

	// sanity check the default values
	if p.self.index != 255 {
		t.Error("p.self.index initialized to ", p.self.index)
	}
	if p.quorum.currentStep != 1 {
		t.Error("s.currentStep should be initialized to 1!")
	}
}

func TestBroadcast(t *testing.T) {
	// tbi
}
