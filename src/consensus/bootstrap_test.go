package consensus

import (
	"network"
	"testing"
)

func TestCreateParticipantFunctions(t *testing.T) {
	rpcs, err := network.NewRPCServer(11000)
	if err != nil {
		t.Fatal(err)
	}

	_, err = CreateBootstrapParticipant(rpcs, "../../filesCreatedDuringTesting/TestCreateParticipantFunctions")
	if err != nil {
		t.Fatal(err)
	}

	// eventually add a few siblings to the quorum, then do status checks to make
	// sure they're all still around, and that balances are changing or something
	// like that.
}
