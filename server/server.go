// Server is the server half of the client-server model that makes up the
// frontend of Sia. The client talks to the server via RPC, and the server runs
// all of the logic that manages participants, wallets, joining the network,
// uploads, etc.
package server

import (
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// Struct Server contains the variables that persist on the server between RPC
// calls. It is the foundation of all operations that require persistence on
// the network.
type Server struct {
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

// NewServer creates a new server object, initializing varibles like maps, and
// processing the configuration file.
func NewServer() (c *Server, err error) {
	// Initialize vital variables.
	c = new(Server)
	c.genericWallets = make(map[GenericWalletID]*GenericWallet)

	// Process config file.
	err = c.processConfigFile()
	if err != nil {
		return
	}

	// more here

	return
}
