package main

import (
	"errors"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (s *Server) broadcast(m network.Message) {
	for i := range s.metadata.Siblings {
		if s.metadata.Siblings[i].Address.Host == "" {
			continue
		}
		m.Dest = s.metadata.Siblings[i].Address
		s.router.SendMessage(m)
		break
	}
}

// Eventually, instead of taking a hostname, there'll be a structure for
// establishing a connection to Sia as a whole, and then finding specific
// quorums within Sia.
func (s *Server) connectToQuorum(connectAddress network.Address) (err error) {
	// Clear out the current metadata and replace it with the new metadata.
	// This is done with an intermediate variable because we do not want to
	// clear the old metadata in the event of an error.
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

// Figure out the latest list of siblings in the quorum.
func (s *Server) refreshMetadata() (err error) {
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
