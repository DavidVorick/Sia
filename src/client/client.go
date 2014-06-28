package client

import (
	"network"
	"participant"
	"quorum"
	"siacrypto"
)

// Struct Client contains the state for client actions
type Client struct {
	router         *network.RPCServer
	genericWallets map[quorum.WalletID]*siacrypto.Keypair
}

// Initializes the client message router and pings the bootstrap to verify
// connectivity.
func (c *Client) Connect(host string, port int) (err error) {
	c.router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}
	// set bootstrap address
	participant.BootstrapAddress.Host = host
	participant.BootstrapAddress.Port = port
	err = c.router.Ping(&participant.BootstrapAddress)
	if err != nil {
		c.router.Close()
	}

	c.siblings[0] = quorum.NewSibling(participant.BootstrapAddress, nil)
	c.RetrieveSiblings()
	return
}

// Get siblings so that each can be uploaded to individually.  This should be
// moved to a (c *Client) function that updates the current siblings. I'm
// actually considering that a client should listen on a quorum, or somehow
// perform lightweight actions (receive digests?) that allow it to keep up but
// don't require many resources.
func (c *Client) RetrieveSiblings() (err error) {
	// Iterate through known siblings until someone provides an updated list. The
	// first answer given is trusted, this is insecure.
	var gobSiblings []byte
	for i := range c.siblings {
		if c.siblings[i] == nil {
			continue
		}
		err = c.router.SendMessage(&network.Message{
			Dest: c.siblings[i].Address(),
			Proc: "Participant.Siblings",
			Args: struct{}{},
			Resp: &gobSiblings,
		})
		if err == nil {
			break
		}
	}
	if err != nil {
		err = fmt.Errorf("Could not reach any stored siblings")
		return
	}

	siblings, err := quorum.DecodeSiblings(gobSiblings)
	if err != nil {
		return
	}
	c.siblings = siblings
	return
}
