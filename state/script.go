package state

import (
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
)

// A ScriptInput pairs an input byte slice with the WalletID associated with
// the recipient. During execution, the WalletID is used to load the script
// body, and then the Input is appended to the end of the script.
type ScriptInput struct {
	Deadline uint32
	Input    []byte
	WalletID WalletID
}

// ScriptInputEvent contains all the information needed by the event list to
// manage the expiration of scripts.
type ScriptInputEvent struct {
	Deadline     uint32
	EventCounter uint32
	Hash         siacrypto.Hash
	WalletID     WalletID
}

func (sie *ScriptInputEvent) Expiration() uint32 {
	return sie.Deadline
}

func (sie *ScriptInputEvent) Counter() uint32 {
	return sie.EventCounter
}

func (sie *ScriptInputEvent) HandleEvent(s *State) (err error) {
	w, err := s.LoadWallet(sie.WalletID)
	if err != nil {
		return
	}

	delete(w.KnownScripts, siafiles.SafeFilename(sie.Hash[:]))

	err = s.SaveWallet(w)
	if err != nil {
		return
	}

	return
}

func (sie *ScriptInputEvent) SetCounter(counter uint32) {
	sie.EventCounter = counter
}

// KnownScript takes the hash of a script as input (it has to be this way,
// unless we move the ScriptInput object to the state package, which is worth
// considering)
func (s *State) KnownScript(si ScriptInput) (known bool, err error) {
	w, err := s.LoadWallet(si.WalletID)
	if err != nil {
		// Here we should really check what type of error got returned.
		// If it's a 'not found' error, then return false. If it's
		// something like a disk-read error, other actions need to be
		// taken.
		known = false
		return
	}

	hash, err := siacrypto.HashObject(si)
	if err != nil {
		return
	}
	_, known = w.KnownScripts[siafiles.SafeFilename(hash[:])]

	return
}

// LearnScript takes a script, creates a script-event using that script, and
// then stores the script in the list of known scripts. When the deadline
// expires, the script event will trigger and the script will be removed.
func (s *State) LearnScript(si ScriptInput) (err error) {
	w, err := s.LoadWallet(si.WalletID)
	if err != nil {
		return
	}

	// Hash the script and add it to the list of known scripts.
	hash, err := siacrypto.HashObject(si)
	if err != nil {
		return
	}

	// Create a script event and add it to the event list.
	sie := ScriptInputEvent{
		Hash:     hash,
		Deadline: si.Deadline,
		WalletID: si.WalletID,
	}

	w.KnownScripts[siafiles.SafeFilename(hash[:])] = sie
	s.InsertEvent(&sie, true)

	err = s.SaveWallet(w)
	if err != nil {
		return
	}

	return
}
