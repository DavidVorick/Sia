package quorum

import (
	"network"
	"testing"
)

// Bootstrap a state to the network, then another
func TestBootstrapping(t *testing.T) {
	// Make a new state and network; start bootstrapping
	z := network.NewZeroNetwork()
	p0, err := CreateParticipant(z)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that s0.self.index updated
	if p0.self.index == 255 {
		t.Error("Bootstrapping failed to update State.self.index")
	}

	// More stuff needs to go here.
	// For now, the testing suite as a whole is being kept lightweight because
	// things are changing rapidly.
}
