package quorum

import (
	"siacrypto"
)

// An event is a task that the quorum will have to perform at a certain block,
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
type event interface {
	handleEvent()
	expiration() uint32
	setCounter(uint64)    // top 32 bits are the expiration, bottom 32 are the counter
	fetchCounter() uint64 // structure will break if fetch does not return the same value called in set
}

// Event nodes have a stack of pointers to the next elements at each height in
// the eventList. The eventNode points to the top, or the furthest reaching
// pointer, and then each pointer points to a less-far reaching pointer. The
// bottom pointer points exactly 1 element forward.
type pointerStack struct {
	nextNode    *eventNode
	nextPointer *pointerStack
}

// An event node houses a pointer to an event, and a pointer to the top of it's
// pointer stack.
type eventNode struct {
	top   *pointerStack
	event event
}

// counts the height of the pointerStack that has been input
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

func (ps *pointerStack) bottom() *pointerStack {
	for ps.nextPointer != nil {
		ps = ps.nextPointer
	}
	return ps
}

func (q *Quorum) insertEvent(e event) {
	// counter has the high 32 bits as the expiration of the event, which allows
	// for sorting according to expiration. Then there's the lower 32 bits which
	// is the eventCounter, and this allows for FCFS unique ordering of events
	// with the same expiration.
	eCounter := uint64(e.expiration())
	eCounter = eCounter << 32
	eCounter += uint64(q.eventCounter)
	q.eventCounter += 1
	e.setCounter(eCounter)
	freshNode := new(eventNode)
	freshNode.event = e

	// check if the current is nil
	if q.eventRoot == nil {
		q.eventRoot = freshNode
		q.eventRoot.event = e
		q.eventRoot.top = new(pointerStack)
		return
	}

	currentHeight := q.eventRoot.top.height()

	// figure out the height of the node to be inserted
	freshHeight := 1
	heightAugmenter, _ := siacrypto.RandomInt(87) // rand from [0, 87)
	for heightAugmenter < 32 {                    // 32/87 is ~ 1/e, the most efficient probability
		freshHeight += 1
		if freshHeight > currentHeight {
			// increase the height of the root node by one
			newTop := new(pointerStack)
			newTop.nextPointer = q.eventRoot.top
			q.eventRoot.top = newTop

			break // root height can only grow by 1 each insertion
		}
		heightAugmenter, _ = siacrypto.RandomInt(87)
	}

	// check if we are behind the root
	currentPointer := q.eventRoot.top
	freshPointer := new(pointerStack)
	if q.eventRoot.event.fetchCounter() >= e.fetchCounter() {
		freshNode.top = q.eventRoot.top
		q.eventRoot.top = freshPointer
		for currentHeight > freshHeight {
			currentPointer = currentPointer.nextPointer
			currentHeight -= 1
		}

		for currentPointer.nextPointer != nil {
			freshPointer.nextPointer = new(pointerStack)
			freshPointer.nextNode = currentPointer.nextNode
			currentPointer.nextNode = q.eventRoot
			freshPointer = freshPointer.nextPointer
			currentPointer = currentPointer.nextPointer
		}
		freshPointer.nextNode = currentPointer.nextNode
		currentPointer.nextNode = q.eventRoot
		q.eventRoot = freshNode
		return
	}

	freshNode.top = freshPointer
	for {
		// move forward until a larger node is found
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

func (q *Quorum) eventNode(e event) *eventNode {
	// check the base cases
	if q.eventRoot == nil {
		return nil
	}
	if q.eventRoot.event.fetchCounter() == e.fetchCounter() {
		return q.eventRoot
	}

	currentPointer := q.eventRoot.top
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

func (q *Quorum) deleteEvent(e event) {
	// first figure out if the event exists by recovering the assiciated eventNode
	en := q.eventNode(e)
	if en == nil {
		return
	}
	if q.eventRoot.top.bottom().nextNode == nil {
		q.eventRoot = nil
		return
	}

	currentHeight := q.eventRoot.top.height()
	currentPointer := q.eventRoot.top
	if q.eventRoot == en {
		// get information on the next node
		currentPointer = currentPointer.bottom()
		nPointer := currentPointer.nextNode.top
		nHeight := nPointer.height()

		currentPointer = q.eventRoot.top
		for currentPointer.nextPointer != nil {
			if currentHeight <= nHeight {
				currentPointer.nextNode = nPointer.nextNode
				nPointer = nPointer.nextPointer
			}
			currentHeight -= 1
			currentPointer = currentPointer.nextPointer
		}

		currentPointer.nextNode.top = q.eventRoot.top
		q.eventRoot = currentPointer.nextNode
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

func (q *Quorum) handleExpiringEvents() {
}
