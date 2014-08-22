package state

import (
	"github.com/NebulousLabs/Sia/siacrypto"
)

const (
	MaxDeadline = 90
)

// A ScriptInput pairs an input byte slice with the WalletID associated with
// the recipient. During execution, the WalletID is used to load the script
// body, and then the Input is appended to the end of the script.
type ScriptInput struct {
	WalletID WalletID
	Input    []byte
	Deadline uint32
}

// ScriptInputEvent contains all the information needed by the event list to
// manage the expiration of scripts.
type ScriptInputEvent struct {
	hash siacrypto.Hash

	// Event-required variables.
	counter    uint32
	expiration uint32
}

func (sie *ScriptInputEvent) Expiration() uint32 {
	return sie.expiration
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
func (s *State) KnownScript(si ScriptInput) bool {
	hash, err := siacrypto.HashObject(si)
	if err != nil {
		// This is a bit of a strange behavior, it might be better to
		// log a panic or something. I'm not sure what would cause a
		// script to be unhashable.
		//
		// I just don't want an errored script getting the okay,
		// because then it will always get the okay.
		return true
	}
	_, exists := s.knownScripts[hash]
	return exists
}

// LearnScript takes a script, creates a script-event using that script, and
// then stores the script in the list of known scripts. When the deadline
// expires, the script event will trigger and the script will be removed.
func (s *State) LearnScript(si ScriptInput) {
	// Hash the script and add it to the list of known scripts.
	hash, err := siacrypto.HashObject(si)
	if err != nil {
		return
	}
	s.knownScripts[hash] = struct{}{}

	// Create a script event and add it to the event list.
	sie := ScriptInputEvent{
		hash:       hash,
		expiration: si.Deadline,
	}

	s.InsertEvent(&sie)
}
