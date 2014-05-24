package quorum

/*
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

	// Verify that s0.self.index updated
	if p0.self.index == 255 {
		t.Error("Bootstrapping failed to update State.self.index")
	}

	// Create a new participant
	p1, err := CreateParticipant(z)
	if err != nil {
		t.Fatal(err)
	}

	// Forward message to bootstrap
	m := z.RecentMessage(0)
	if m == nil {
		t.Fatal("message 0 never received")
	} else if m.Proc != "Participant.JoinSia" {
		t.Fatal("message 0 has wrong type: expected Participant.JoinSia, got", m.Proc)
	}
	p0.JoinSia(m.Args.(Sibling), nil)

	// Verify that a broadcast message went out indicating a new sibling
	m = z.RecentMessage(1)
	if m == nil {
		t.Fatal("message 1 never received")
	} else if m.Proc != "Participant.AddHopeful" {
		t.Fatal("message 1 has wrong type: expected Participant.AddHopeful, got", m.Proc)
	}

	// skip ahead and just add the new sibling
	s := m.Args.(Sibling)
	s.index = 1
	p0.addNewSibling(&s)

	// deliver download message to new sibling
	gobQuorum, _ := p0.quorum.GobEncode()
	p1.TransferQuorum(gobQuorum, nil)

	// ticking is not being tested at the moment, due to a race condition on
	// the locks. The test gets the locks before the tick() function does, which
	// means that ticking is not updated by the time the test is checking it.

	// additionally, no testing is done to see that the siblings are sending
	// messages to each other.
}
*/
