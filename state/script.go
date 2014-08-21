package state

import (
	"github.com/NebulousLabs/Sia/siacrypto"
)

// ScriptInputEvent contains all the information needed by the event list to
// manage the expiration of scripts.
type ScriptInputEvent struct {
	hash siacrypto.Hash

	// Event-required variables.
	counter  uint32
	deadline uint32
}

func (sie *ScriptInputEvent) Expiration() uint32 {
	return sie.deadline
}

func (sie *ScriptInputEvent) Counter() uint32 {
	return sie.counter
}

func (sie *ScriptInputEvent) HandleEvent(s *State) {
	delete(s.knownScripts, sie.hash)
}

func (sie *ScriptInputEvent) SetCounter(counter uint32) {
	sie.counter = counter
}

// KnownScript takes the hash of a script as input (it has to be this way,
// unless we move the ScriptInput object to the state package, which is worth
// considering)
func (s *State) KnownScript() bool {
	// to be implemented after delta.ScriptInput is transitioned to state.ScriptInput
	return false
}

// LearnScript takes a script, creates a script-event using that script, and
// then stores the script in the list of known scripts. When the deadline
// expires, the script event will trigger and the script will be removed.
func (s *State) LearnScript() {
	// to be implemented after delta.ScriptInput is transitioned to state.ScriptInput
}
