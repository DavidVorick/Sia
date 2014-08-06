package client

import (
	"errors"
	"network"
)

// The Server houses all of the participants. It contains a single message
// router that is shared by all of the participants, it will eventually contain
// a clock object that will be used and modified by all participants.
type Server struct {
	networkServer *network.RPCServer
}

// NewServer takes a port number as input and returns a server object that's
// ready to be populated with participants.
func (c *Client) NewServer(port int) (err error) {
	// Prevent any existing server from being overwritten.
	if c.participantServer != nil {
		err = errors.New("server already exists")
		return
	}

	// Create the RPCServer.
	networkServer, err := network.NewRPCServer(port)
	if err != nil {
		return
	}

	// Establish c.participantServer.
	c.participantServer = new(Server)
	c.participantServer.networkServer = networkServer
	return
}
