package client

import (
	"errors"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

// Initializes the client message router and pings the bootstrap to verify
// connectivity.
func (c *Client) BootstrapConnection(connectAddress network.Address) (err error) {
	// A network server needs to exist in order to connect to the network.
	if c.router == nil {
		err = errors.New("network router is nil")
		return
	}

	// Ping the connectAddress to verify alive-ness.
	err = c.router.Ping(connectAddress)
	if err != nil {
		return
	}

	// Discover our external IP.
	// Eventually, LearnHostname will ask the boostrap address instead of an
	// external service.
	err = c.router.LearnHostname()
	if err != nil {
		return errors.New("failed to determine external IP")
	}

	// populate initial sibling list
	var metadata state.Metadata
	err = c.router.SendMessage(network.Message{
		Dest: connectAddress,
		Proc: "Participant.Metadata",
		Args: struct{}{},
		Resp: &metadata,
	})
	if err != nil {
		return
	}
	c.metadata.Siblings = metadata.Siblings
	return
}

// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (c *Client) Broadcast(m network.Message) {
	for i := range c.metadata.Siblings {
		if c.metadata.Siblings[i].Address.Host == "" {
			continue
		}
		m.Dest = c.metadata.Siblings[i].Address
		c.router.SendMessage(m)
		break
	}
}

// Connect will create a new router, listening on the input port. If there is
// already a non-nil router in the client, an error will be returned.
func (c *Client) Connect(port uint16) (err error) {
	if c.router != nil {
		err = errors.New("router already initialized")
		return
	}

	c.router, err = network.NewRPCServer(port)
	if err != nil {
		return
	}
	return
}

// IsRouterInitialized is useful for telling front end programs whether a
// router needs to be initialized or not.
func (c *Client) IsRouterInitialized() bool {
	return c.router != nil
}

// Figure out the latest list of siblings in the quorum.
func (c *Client) RefreshMetadata() (err error) {
	// Iterate through known siblings until someone provides an updated list.
	// The first answer given is trusted, this is insecure. A separate
	// variable, 'metadata', is used instead of 'c.metadata' because an
	// erroneous call may wipe out the metadata otherwise.
	var metadata state.Metadata
	for i := range c.metadata.Siblings {
		if c.metadata.Siblings[i].Address.Host == "" {
			continue
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.metadata.Siblings[i].Address,
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

	c.metadata = metadata

	return
}
