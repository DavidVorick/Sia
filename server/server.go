// Server is the server half of the client-server model that makes up the
// frontend of Sia. The client talks to the server via RPC, and the server runs
// all of the logic that manages participants, wallets, joining the network,
// uploads, etc.
package main

import (
	"fmt"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// Struct Server contains the variables that persist on the server between RPC
// calls. It is the foundation of all operations that require persistence on
// the network.
type Server struct {
	// Networking Variables
	router   *network.RPCServer
	address  network.Address
	metadata state.Metadata

	// Generic Wallets
	// A pointer to the generic wallet type is stored because we wish to
	// pass and manipulate the generic wallet by reference. Maps are not
	// pointer safe - you can't pass a pointer to an object in the map.
	genericWallets map[GenericWalletID]*GenericWallet

	participantManager *ParticipantManager
}

// connect creates a router for the server, learning a public hostname if the
// flag is set.
func (s *Server) connect(port uint16, learnHostname bool) (err error) {
	// Create a router.
	s.router, err = network.NewRPCServer(port)
	if err != nil {
		return
	}

	// Register with the router, calling LearnHostname() if we wish to be
	// available to the public.
	if learnHostname {
		err = s.router.LearnHostname()
		if err != nil {
			return
		}
	}
	s.address = s.router.RegisterHandler(s)

	// Create a participant manager.
	s.participantManager, err = newParticipantManager()
	if err != nil {
		fmt.Printf("Participant Manager Error: %v!\n", err)
		err = nil
	}

	return
}

// newServer creates a server struct, initializing key variables like maps.
// Note that it does not initialize the router.
func newServer() (s *Server) {
	s = new(Server)
	s.genericWallets = make(map[GenericWalletID]*GenericWallet)
	return
}
