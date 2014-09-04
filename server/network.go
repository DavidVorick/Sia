package main

import (
	"errors"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// Initializes the client message router and pings the bootstrap to verify
// connectivity.
func (s *Server) BootstrapConnection(connectAddress network.Address) (err error) {
	// A network server needs to exist in order to connect to the network.
	if s.router == nil {
		err = errors.New("network router is nil")
		return
	}

	// Ping the connectAddress to verify alive-ness.
	err = s.router.Ping(connectAddress)
	if err != nil {
		return
	}

	// Discover our external IP.
	// Eventually, LearnHostname will ask the boostrap address instead of an
	// external service.
	err = s.router.LearnHostname()
	if err != nil {
		return errors.New("failed to determine external IP")
	}

	// populate initial sibling list
	var metadata state.Metadata
	err = s.router.SendMessage(network.Message{
		Dest: connectAddress,
		Proc: "Participant.Metadata",
		Args: struct{}{},
		Resp: &metadata,
	})
	if err != nil {
		return
	}
	s.metadata.Siblings = metadata.Siblings
	return
}

// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (s *Server) Broadcast(m network.Message) {
	for i := range s.metadata.Siblings {
		if s.metadata.Siblings[i].Address.Host == "" {
			continue
		}
		m.Dest = s.metadata.Siblings[i].Address
		s.router.SendMessage(m)
		break
	}
}

// Connect will create a new router, listening on the input port. If there is
// already a non-nil router in the client, an error will be returned.
func (s *Server) Connect(port uint16) (err error) {
	if s.router != nil {
		err = errors.New("router already initialized")
		return
	}

	s.router, err = network.NewRPCServer(port)
	if err != nil {
		return
	}
	return
}

// IsRouterInitialized is useful for telling front end programs whether a
// router needs to be initialized or not.
func (s *Server) IsRouterInitialized() bool {
	return s.router != nil
}

// Figure out the latest list of siblings in the quorum.
func (s *Server) RefreshMetadata() (err error) {
	// Iterate through known siblings until someone provides an updated list.
	// The first answer given is trusted, this is insecure. A separate
	// variable, 'metadata', is used instead of 's.metadata' because an
	// erroneous call may wipe out the metadata otherwise.
	var metadata state.Metadata
	for i := range s.metadata.Siblings {
		if s.metadata.Siblings[i].Address.Host == "" {
			continue
		}
		err = s.router.SendMessage(network.Message{
			Dest: s.metadata.Siblings[i].Address,
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &metadata,
		})
		if err == nil {
			// End the loop after first success.
			break
		}
	}
	if err != nil {
		err = errors.New("Could not reach any stored siblings")
		return
	}

	s.metadata = metadata

	return
}
