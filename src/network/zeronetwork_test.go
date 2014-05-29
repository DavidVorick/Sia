package network

import (
	"testing"
)

// use all functions of ZeroNetwork, expecting no nil returns and no errors
func TestZeroNetwork(t *testing.T) {
	zeroNet := NewZeroNetwork()
	if zeroNet == nil {
		t.Fatal("NewZeroNetwork cannot return nil")
	}

	m := &Message{
		zeroNet.Address(),
		"",
		nil,
		nil,
	}

	err := zeroNet.SendMessage(m)
	if err != nil {
		t.Fatal("ZeroNetwork.SendMessage cannot fail")
	}
}
