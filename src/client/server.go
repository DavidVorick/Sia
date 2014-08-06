package client

import (
	"consensus"
	"errors"
	"network"
)

// The Server houses all of the participants. It contains a single message
// router that is shared by all of the participants, it will eventually contain
// a clock object that will be used and modified by all participants.
type Server struct {
	networkServer *network.RPCServer

	participants map[network.Identifier]consensus.Participant
}

// NewServer takes a port number as input and returns a server object that's
// ready to be populated with participants.
func (c *Client) NewServer() (err error) {
	// If the network router is nil, a server can't exist.
	if c.router == nil {
		err = errors.New("need to have a connection before creating a server")
		return
	}

	// Prevent any existing server from being overwritten.
	if c.participantServer != nil {
		err = errors.New("server already exists")
		return
	}

	// Establish c.participantServer.
	c.participantServer = new(Server)
	return
}
