package state

import (
	"siacrypto"
)

// An expired script is an event
type ExpiredScript struct {
	ScriptHash siacrypto.Hash
}

func (es *ExpiredScript) handleEvent(s *State) {
	//
}
