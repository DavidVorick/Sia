package quorum

import (
	"common"
	"testing"
)

// Bootstrap a state to the network, then another
func TestJoinQuorum(t *testing.T) {
	// Make a new state and network; start bootstrapping
	z := common.NewZeroNetwork()
	p0, err := CreateParticipant(z)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the message for correctness

	// Forward message to bootstrap State (ourselves, as it were)
	m := z.RecentMessage(0)
	if m == nil {
		t.Fatal("message 0 never received")
	}
	p0.JoinSia(m.Args.(Sibling), nil)

	// Verify that a broadcast message went out indicating a new sibling

	// Forward message to recipient
	m = z.RecentMessage(1)
	if m == nil {
		t.Fatal("message 1 never received")
	}
	p0.AddNewSibling(m.Args.(Sibling), nil)

	// Verify that we started ticking
	p0.quorum.tickingLock.Lock()
	if !p0.quorum.ticking {
		t.Fatal("Bootstrap state not ticking after joining Sia")
	}
	p0.quorum.tickingLock.Unlock()

	// Verify that s0.self.index updated
	if p0.self.index == 255 {
		t.Error("Bootstrapping failed to update State.self.index")
	}

	// Create a new state to bootstrap
	p1, err := CreateParticipant(z)
	if err != nil {
		t.Fatal(err)
	}

	// Verify message for correctness

	// Deliver message to bootstrap
	m = z.RecentMessage(2)
	p0.JoinSia(m.Args.(Sibling), nil)

	// Deliver the broadcasted messages
	m = z.RecentMessage(3)
	p0.AddNewSibling(m.Args.(Sibling), nil)
	m = z.RecentMessage(4)
	p1.AddNewSibling(m.Args.(Sibling), nil)

	// Verify the messages made it
	p1.quorum.tickingLock.Lock()
	if !p1.quorum.ticking {
		t.Error("s1 did not start ticking")
	}

	// both swarms should be aware of each other... maybe test their ongoing interactions?
}
