package quorum

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

func (q *Quorum) insertEvent(e event) {
}

func (q *Quorum) deleteEvent(e event) {
}

func (q *Quorum) handleExpiringEvents() {
}
