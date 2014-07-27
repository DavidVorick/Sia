package state

const (
	MaxDeadline = 300
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
type EventInterface interface {
	handleEvent(s *State)
}

// The Event type is a generic type that is meant to be switched upon. Each
// event has its own set of functions but they all go into a single event list
// together.
type Event struct {
	Type         string
	Expiration   uint32
	Counter      uint32
	EncodedEvent []byte
}
