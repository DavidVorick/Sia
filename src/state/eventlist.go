package state

import (
	"siacrypto"
)

// An eventNode houses an event and a pointer to the top of its pointer stack.
// The pointer stack is the structure used to implement a skip list.
type eventNode struct {
	top   *pointerStack
	event EventInterface
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

// height() returns the number of elements in the linked list of pointerStacks,
// including the one used to call height().
func (ps *pointerStack) height() (i int) {
	if ps != nil {
		i += 1
	}
	for ps.nextPointer != nil {
		i += 1
		ps = ps.nextPointer
	}
	return
}

// bottom() takes a pointerStack and finds the bottom pointerStack, which is
// guaranteed to point only one event forward.
func (ps *pointerStack) bottom() *pointerStack {
	for ps.nextPointer != nil {
		ps = ps.nextPointer
	}
	return ps
}

// InsertEvent takes a new event and inserts it into the event list.
// InsertEvent does not check that the event already exists inside of the list,
// as duplicate events are allowed to exist in the event list.
func (s *State) InsertEvent(e EventInterface) {
	// counter has the high 32 bits as the expiration of the event, which allows
	// for sorting according to expiration. Then there's the lower 32 bits which
	// is the eventCounter, and this allows for FCFS unique ordering of events
	// with the same expiration.
	eCounter := uint64(e.expiration())
	eCounter = eCounter << 32
	eCounter += uint64(s.Metadata.EventCounter)
	s.Metadata.EventCounter += 1
	e.setCounter(eCounter)
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
	heightAugmenter, _ := siacrypto.RandomInt(87) // rand from [0, 87)
	for heightAugmenter < 32 {                    // 32/87 is ~ 1/e, the most efficient probability
		freshHeight += 1
		if freshHeight > currentHeight {
			// increase the height of the root node by one
			newTop := new(pointerStack)
			newTop.nextPointer = s.eventRoot.top
			s.eventRoot.top = newTop

			break // root height can only grow by 1 each insertion
		}
		heightAugmenter, _ = siacrypto.RandomInt(87)
	}

	// check if we are behind the root
	currentPointer := s.eventRoot.top
	freshPointer := new(pointerStack)
	if s.eventRoot.event.fetchCounter() >= e.fetchCounter() {
		freshNode.top = s.eventRoot.top
		s.eventRoot.top = freshPointer
		for currentHeight > freshHeight {
			currentPointer = currentPointer.nextPointer
			currentHeight -= 1
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
		for currentPointer.nextNode != nil && currentPointer.nextNode.event.fetchCounter() < e.fetchCounter() {
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
		currentHeight -= 1
	}
}

// deleteEvent takes an event, finds it in the event list, and then deletes the
// event. This is called after an event expires, and is also called if an event
// is triggered before its expiration.
func (s *State) DeleteEvent(e EventInterface) {
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
			currentHeight -= 1
			currentPointer = currentPointer.nextPointer
		}

		currentPointer.nextNode.top = s.eventRoot.top
		s.eventRoot = currentPointer.nextNode
		currentPointer.nextNode = nPointer.nextNode
		return
	}

	// then go through finding the event, repointing everything that points to the event as a height object
	eventHeight := en.top.height()
	eventPointer := en.top
	for {
		// move forward
		for currentPointer.nextNode != nil && currentPointer.nextNode.event.fetchCounter() < e.fetchCounter() {
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
		currentHeight -= 1
	}
}

// eventNode() retreives an eventNode that corresponds to a specific event.
// eventNode() is only used internally.
func (s *State) eventNode(e EventInterface) *eventNode {
	// check the base cases
	if s.eventRoot == nil {
		return nil
	}
	if s.eventRoot.event.fetchCounter() == e.fetchCounter() {
		return s.eventRoot
	}

	currentPointer := s.eventRoot.top
	for {
		// move forward
		for currentPointer.nextNode != nil && currentPointer.nextNode.event.fetchCounter() < e.fetchCounter() {
			currentPointer = currentPointer.nextNode.top
		}

		// see if the next node is the desired node
		if currentPointer.nextNode != nil && currentPointer.nextNode.event.fetchCounter() == e.fetchCounter() {
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
