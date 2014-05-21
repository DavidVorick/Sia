package quorum

import (
	"common"
	"testing"
)

// Create a state, check the defaults
func TestCreateParticipant(t *testing.T) {
	zn := common.NewZeroNetwork()
	// make sure CreateState does not cause errors
	p0, err := CreateParticipant(zn)
	if err != nil {
		t.Fatal(err)
	}

	// sanity check the default values for the bootstrap
	if p0.self.index != 0 {
		t.Error("p0.self.index initialized to", p0.self.index)
	}
	if p0.currentStep != 1 {
		t.Error("p0.currentStep should be initialized to 1!")
	}

	// check a non-bootstrap
	p1, err := CreateParticipant(zn)
	if err != nil {
		t.Fatal(err)
	}
	if p1.self.index != 255 {
		t.Error("p1.self.index initialized to", p1.self.index)
	}
	if p1.currentStep != 1 {
		t.Error("p1.currentStep should be initialized to 1!")
	}
}

func TestBroadcast(t *testing.T) {
	// tbi
}
