package network

import (
	"testing"
)

// use all functions of DebugNetwork, expecting no nil returns and no errors
func TestDebugNetwork(t *testing.T) {
	debugNet := NewDebugNetwork()
	if debugNet == nil {
		t.Fatal("NewDebugNetwork cannot return nil")
	}

	m := &Message{
		debugNet.Address(),
		"",
		nil,
		nil,
	}

	err := debugNet.SendMessage(m)
	if err != nil {
		t.Fatal("DebugNetwork.SendMessage cannot fail")
	}
}
