package state

import (
	"testing"

	"github.com/NebulousLabs/Sia/siacrypto"
)

// countReachableEvents starts at the event root and figures out how many are
// reachable from a bottom-level crawl of the event list. Only checks for
// infinite loops where an event points to itself.
func countReachableEvents(en *eventNode) (i int) {
	for {
		if en == nil {
			return
		}
		i += 1

		current := en.top
		if current == nil {
			return
		}
		current = current.bottom()
		if current.nextNode == en {
			panic("countReachableEvents has detected an infinite loop")
		}
		en = current.nextNode
	}
}

// TestEventList is designed to verify that the skip-list logic of the event
// list is reasonably responsive and doesn't have any unexpected behaviors,
// such as failing to remove an event after calling delte.
//
// A current shortcoming of this test is that it doesn't check for ordering,
// and Process() is also not verified against.
func TestEventList(t *testing.T) {
	// Create and initialize a state.
	var s State
	s.Initialize()

	// Create and insert an event.
	e0 := &ScriptInputEvent{
		expiration: 1,
	}
	s.InsertEvent(e0)

	en0 := s.eventNode(e0)
	if en0 == nil {
		t.Fatal("Could not get inserted event!")
	}
	if countReachableEvents(s.eventRoot) != 1 {
		t.Fatal("Reached wrong number of events, expecting 1:", countReachableEvents(s.eventRoot))
	}

	s.DeleteEvent(e0)
	en0 = s.eventNode(e0)
	if en0 != nil {
		t.Fatal("deleted event node still retrievable")
	}
	if countReachableEvents(s.eventRoot) != 0 {
		t.Fatal("Reached wrong number of events, expecting 0:", countReachableEvents(s.eventRoot))
	}

	if testing.Short() {
		t.Skip()
	}

	sieMap := make(map[*ScriptInputEvent]struct{})
	n := 50
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			randomTimeout := siacrypto.RandomInt(12)
			si := &ScriptInputEvent{
				expiration: uint32(randomTimeout),
			}
			sieMap[si] = struct{}{}
			s.InsertEvent(si)

			if countReachableEvents(s.eventRoot) != i+1 {
				t.Error("Reached wrong number of events, expecting", i+1, "got", countReachableEvents(s.eventRoot))
			}
		}

		elementSlice := make([]*ScriptInputEvent, n)
		i := 0
		for key := range sieMap {
			elementSlice[i] = key
			i++
		}

		// try and fetch every element
		for i := range elementSlice {
			wn := s.eventNode(elementSlice[i])
			if wn == nil {
				t.Error("cannot reach inserted element")
			}
		}

		// shuffle elementSlice
		for i := range elementSlice {
			newIndex := siacrypto.RandomInt(len(elementSlice) - i)
			newIndex += i

			tmp := elementSlice[newIndex]
			elementSlice[newIndex] = elementSlice[i]
			elementSlice[i] = tmp
		}

		for i := range elementSlice {
			s.DeleteEvent(elementSlice[i])
			wn := s.eventNode(elementSlice[i])
			if wn != nil {
				t.Error("deleted event node is still fetchable")
			}
			if countReachableEvents(s.eventRoot) != n-i-1 {
				t.Fatal("Wrong number of reachable events, expecting", n-i-1, "got", countReachableEvents(s.eventRoot))
			}
		}
		sieMap = make(map[*ScriptInputEvent]struct{})
	}

	// insert a bunch of random things
	// randomly insert and delete the things
	// delete all of the things
	// check sorting each time
}
