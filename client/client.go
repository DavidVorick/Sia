package client

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
	genericWallets map[state.WalletID]GenericWallet

	// Participant Server
	participantServer *Server
}

// Uses the configuration file to create a new client, initializing variables
// and connecting to the network as specified by the configuration.
func NewClient() (c *Client, err error) {
	// Initialize vital variables.
	c = new(Client)
	c.genericWallets = make(map[state.WalletID]GenericWallet)

	// Process config file.
	err = c.processConfigFile()
	if err != nil {
		return
	}

	// more here

	return
}
