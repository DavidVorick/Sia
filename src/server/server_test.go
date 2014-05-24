package main

/*
import (
	"common"
	"network"
	"quorum"
	"testing"
	"time"
)

func TestNetworkedQuorum(t *testing.T) {
	// create a MessageRouter and 4 participants
	rpcs, err := network.NewRPCServer(9988)
	if err != nil {
		println("message sender creation failed")
	}

	_, err = quorum.CreateParticipant(rpcs)
	if err != nil {
		println("p0 creation failed")
	}
	_, err = quorum.CreateParticipant(rpcs)
	if err != nil {
		println("p1 creation failed")
	}
	_, err = quorum.CreateParticipant(rpcs)
	if err != nil {
		println("p2 creation failed")
	}
	_, err = quorum.CreateParticipant(rpcs)
	if err != nil {
		println("p3 creation failed")
	}

	// Basically checking for errors up to this point
	if testing.Short() {
		t.Skip()
	}

	time.Sleep(3 * common.StepDuration * time.Duration(common.QuorumSize))

	// if no seg faults, no errors
	// there needs to be a s0.ParticipantStatus() call returning a function with public information about the participant
	// there needs to be a s0.QuorumStatus() call returning public information about the quorum
	// 		all participants in a public quorum should return the same information
}*/
