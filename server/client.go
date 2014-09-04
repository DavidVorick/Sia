package server

import (
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// Struct Client contains the state for the client, the variables that persist
// between function calls.
type Client struct {
	// Networking Variables
	router   *network.RPCServer
	metadata state.Metadata

	// Generic Wallets
	// A pointer to the generic wallet type is stored because we wish to
	// pass and manipulate the generic wallet by reference. Maps are not
	// pointer safe - you can't pass a pointer to an object in the map.
	genericWallets map[GenericWalletID]*GenericWallet

	participantManager *ParticipantManager
}

// Uses the configuration file to create a new client, initializing variables
// and connecting to the network as specified by the configuration.
func NewClient() (c *Client, err error) {
	// Initialize vital variables.
	c = new(Client)
	c.genericWallets = make(map[GenericWalletID]*GenericWallet)

	// Process config file.
	err = c.processConfigFile()
	if err != nil {
		return
	}

	// more here

	return
}
