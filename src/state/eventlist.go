package state

import (
	"siacrypto"
)

const (
	// MaxDeadline is the maximum allowed expiration deadline for Events.
	MaxDeadline = 300
)

// An Event is a task that the quorum will have to perform at a certain block,
// which is returned by expiration(). Something may trigger the event early, at
// which point the event will be deleted from the eventList. Each block, all
// events that expire that block are handled by calling handleEvent() on the
// event. They are then removed from the eventList. The event list keeps all
// events in order of expiration, meaning that you only need to check the
// beginning of the eventList until an event is found that expires at a later
// block. The internals of the event list are determined randomly and
// nondeterministically, because the internals do not need to be consistent
// between siblings. This also prevents an attacker from knowing the internals
// and being able to provide malicious input to distrupt the order notation of
// the list.
type Event interface {
	Expiration() uint32
	Counter() uint32
	SetCounter(uint32)
	HandleEvent(s *State)
}

// An eventNode houses an event and a pointer to the top of its pointer stack.
// The pointer stack is the structure used to implement a skip list.
type eventNode struct {
	top   *pointerStack
	event Event
}

// A pointerStack is a linked list of pointers to the next elements at each
// height of the eventList. The eventNode points to the top pointer, which is
// the pointer that reaches the farthest. Each pointer below that points to a
// closer element, until you get to the bottom pointerStack, which points
// exactly 1 element forward. The field 'nextPointer' is nil for the lowest
// pointerStack.
type pointerStack struct {
	nextNode    *eventNode
	nextPointer *pointerStack
}

// height returns the number of elements in the linked list of pointerStacks,
// including the one used to call height().
func (ps *pointerStack) height() (i int) {
	if ps != nil {
		i++
	}
	for ps.nextPointer != nil {
		i++
		ps = ps.nextPointer
	}
	return
}

// bottom takes a pointerStack and finds the bottom pointerStack, which is
// guaranteed to point only one event forward.
func (ps *pointerStack) bottom() *pointerStack {
	for ps.nextPointer != nil {
		ps = ps.nextPointer
	}
	return ps
}

func eventIndex(e Event) (index uint64) {
	index = uint64(e.Expiration()) << 32
	index += uint64(e.Counter())
	return
}

// InsertEvent takes a new event and inserts it into the event list.
// It does not check that the event already exists inside of the list,
// as duplicate events are allowed to exist in the event list.
func (s *State) InsertEvent(e Event) {
	// counter has the high 32 bits as the expiration of the event, which allows
	// for sorting according to expiration. Then there's the lower 32 bits which
	// is the eventCounter, and this allows for FCFS unique ordering of events
	// with the same expiration.
	e.SetCounter(s.Metadata.EventCounter)
	s.Metadata.EventCounter++
	freshNode := new(eventNode)
	freshNode.event = e

	// check if the current is nil
	if s.eventRoot == nil {
		s.eventRoot = freshNode
		s.eventRoot.event = e
		s.eventRoot.top = new(pointerStack)
		return
	}

	currentHeight := s.eventRoot.top.height()

	// Determine the height of the node to be inserted. This height is decided
	// individually by each participant, and does not need to be the same across
	// quorums. Having every event list use a different implementation makes it
	// much harder to attack the skip list by inserting events in a malicious
	// order.
	freshHeight := 1
	heightAugmenter := siacrypto.RandomInt(87) // rand from [0, 87)
	for heightAugmenter < 32 {                 // 32/87 is ~ 1/e, the most efficient probability
		freshHeight++
		if freshHeight > currentHeight {
			// increase the height of the root node by one
			newTop := new(pointerStack)
			newTop.nextPointer = s.eventRoot.top
			s.eventRoot.top = newTop

			break // root height can only grow by 1 each insertion
		}
		heightAugmenter = siacrypto.RandomInt(87)
	}

	// check if we are behind the root
	currentPointer := s.eventRoot.top
	freshPointer := new(pointerStack)
	if eventIndex(s.eventRoot.event) >= eventIndex(e) {
		freshNode.top = s.eventRoot.top
		s.eventRoot.top = freshPointer
		for currentHeight > freshHeight {
			currentPointer = currentPointer.nextPointer
			currentHeight--
		}

		for currentPointer.nextPointer != nil {
			freshPointer.nextPointer = new(pointerStack)
			freshPointer.nextNode = currentPointer.nextNode
			currentPointer.nextNode = s.eventRoot
			freshPointer = freshPointer.nextPointer
			currentPointer = currentPointer.nextPointer
		}
		freshPointer.nextNode = currentPointer.nextNode
		currentPointer.nextNode = s.eventRoot
		s.eventRoot = freshNode
		return
	}

	freshNode.top = freshPointer
	for {
		// Move forward until a larger node is found.
		for currentPointer.nextNode != nil && eventIndex(currentPointer.nextNode.event) < eventIndex(e) {
			currentPointer = currentPointer.nextNode.top
		}

		// update pointer if needed
		if currentHeight <= freshHeight {
			freshPointer.nextNode = currentPointer.nextNode
			currentPointer.nextNode = freshNode

			// break the loop if we're at the bottom of the list. This logic will
			// always be reached if we are at the bottom of the list, because at the
			// bottom of the list currentHeight will be 1, and freshHeight will never
			// be less than 1.
			if currentPointer.nextPointer == nil {
				break
			}
			freshPointer.nextPointer = new(pointerStack)
			freshPointer = freshPointer.nextPointer
		}

		// move down
		currentPointer = currentPointer.nextPointer
		currentHeight--
	}
}

// DeleteEvent takes an event, finds it in the event list, and then deletes the
// event. This is called after an event expires, and is also called if an event
// is triggered before its expiration.
func (s *State) DeleteEvent(e Event) {
	// first figure out if the event exists by recovering the assiciated eventNode
	en := s.eventNode(e)
	if en == nil {
		return
	}
	if s.eventRoot.top.bottom().nextNode == nil {
		s.eventRoot = nil
		return
	}

	currentHeight := s.eventRoot.top.height()
	currentPointer := s.eventRoot.top
	if s.eventRoot == en {
		// get information on the next node
		currentPointer = currentPointer.bottom()
		nPointer := currentPointer.nextNode.top
		nHeight := nPointer.height()

		currentPointer = s.eventRoot.top
		for currentPointer.nextPointer != nil {
			if currentHeight <= nHeight {
				currentPointer.nextNode = nPointer.nextNode
				nPointer = nPointer.nextPointer
			}
			currentHeight--
			currentPointer = currentPointer.nextPointer
		}

		currentPointer.nextNode.top = s.eventRoot.top
		s.eventRoot = currentPointer.nextNode
		currentPointer.nextNode = nPointer.nextNode
		return
	}

	// then go through finding the event, repointing everything that points to
	// the event as a height object
	eventHeight := en.top.height()
	eventPointer := en.top
	for {
		// move forward
		for currentPointer.nextNode != nil && eventIndex(currentPointer.nextNode.event) < eventIndex(e) {
			currentPointer = currentPointer.nextNode.top
		}

		if eventHeight >= currentHeight {
			currentPointer.nextNode = eventPointer.nextNode
			eventPointer = eventPointer.nextPointer

			if currentPointer.nextPointer == nil {
				break
			}
		}

		// move down
		currentPointer = currentPointer.nextPointer
		currentHeight--
	}
}

// eventNode() retreives an eventNode that corresponds to a specific event.
func (s *State) eventNode(e Event) *eventNode {
	// check the base cases
	if s.eventRoot == nil {
		return nil
	}
	if eventIndex(s.eventRoot.event) == eventIndex(e) {
		return s.eventRoot
	}

	currentPointer := s.eventRoot.top
	for {
		// move forward
		for currentPointer.nextNode != nil && eventIndex(currentPointer.nextNode.event) < eventIndex(e) {
			currentPointer = currentPointer.nextNode.top
		}

		// see if the next node is the desired node
		if currentPointer.nextNode != nil && eventIndex(currentPointer.nextNode.event) == eventIndex(e) {
			return currentPointer.nextNode
		}

		// see if we're at the bottom of the list - node not found
		if currentPointer.nextPointer == nil {
			return nil
		}

		// move down
		currentPointer = currentPointer.nextPointer
	}
}

/*
func (q *Quorum) ProcessEvents() {
	for q.eventRoot != nil && q.eventRoot.event.expiration() <= q.height {
		q.eventRoot.event.handleEvent(q)
		q.deleteEvent(q.eventRoot.event)
	}
}*/
