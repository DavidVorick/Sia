package participant

import (
	"network"
	"testing"
)

// somewhere in this file we need a test that gob.Register is being called for
// all updates

func TestSynchronizeEncoding(t *testing.T) {
	// tbi
}

// Create a state, check the defaults
func TestCreateParticipant(t *testing.T) {
	zn := network.NewDebugNetwork()
	// make sure CreateState does not cause errors
	p0, err := CreateParticipant(zn)
	if err != nil {
		t.Fatal(err)
	}

	// sanity check the default values for the bootstrap
	if p0.self.Index() != 0 {
		t.Error("p0.self.index initialized to", p0.self.Index())
	}
	if p0.currentStep != 1 {
		t.Error("p0.currentStep should be initialized to 1!")
	}

	// check a non-bootstrap
	p1, err := CreateParticipant(zn)
	if err != nil {
		t.Fatal(err)
	}
	if p1.self.Index() != 255 {
		t.Error("p1.self.index initialized to", p1.self.Index())
	}

	// test creating another participant that doesn't have the bootstrap address
}
