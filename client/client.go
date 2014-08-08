package client

import (
	"errors"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

type Keypair struct {
	PK siacrypto.PublicKey
	SK siacrypto.SecretKey
}

// Struct Client contains the state for client actions
type Client struct {
	// Networking Variables
	router    *network.RPCServer
	bootstrap network.Address

	// State Variables
	siblings [state.QuorumSize]state.Sibling

	// All Wallets
	genericWallets map[state.WalletID]Keypair
}

/*
// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (c *Client) Broadcast(nm network.Message) {
	for i := range c.siblings {
		if c.siblings[i].Address.Host == "" {
			continue
		}
		nm.Dest = c.siblings[i].Address
		c.router.SendMessage(nm)
		break
	}
}
*/

// Initializes the client message router and pings the bootstrap to verify
// connectivity.
func (c *Client) Connect(host string, port uint16, id byte) (err error) {
	c.router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}

	// Set the bootstrap address.
	c.bootstrap = network.Address{
		Host: host,
		Port: port,
		ID:   network.Identifier(id),
	}

	// Check that the address is valid.
	err = c.router.Ping(c.bootstrap)
	if err != nil {
		c.router.Close()
		return
	}

	// populate initial sibling list
	var metadata state.Metadata
	err = c.router.SendMessage(network.Message{
		Dest: c.bootstrap,
		Proc: "Participant.Metadata",
		Args: struct{}{},
		Resp: &metadata,
	})
	if err != nil {
		c.router.Close()
		return
	}
	c.siblings = metadata.Siblings
	return
}

// Closes and destroys the client's RPC server
func (c *Client) Disconnect() {
	if c.router == nil {
		return
	}
	c.router.Close()
	c.router = nil
}

// Get siblings so that each can be uploaded to individually.  This should be
// moved to a (c *Client) function that updates the current siblings. I'm
// actually considering that a client should listen on a quorum, or somehow
// perform lightweight actions (receive digests?) that allow it to keep up but
// don't require many resources.
func (c *Client) RetrieveSiblings() (err error) {
	// Iterate through known siblings until someone provides an updated list. The
	// first answer given is trusted, this is insecure.
	var metadata state.Metadata
	for i := range c.siblings {
		if c.siblings[i].Address.Host == "" {
			continue
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.siblings[i].Address,
			Proc: "Participant.Metadata",
			Args: struct{}{},
			Resp: &metadata,
		})
		if err == nil {
			break
		}
		c.siblings = metadata.Siblings
	}
	if err != nil {
		err = errors.New("Could not reach any stored siblings")
		return
	}

	return
}

// Creates a client, follows the instructions of the config file, and returns a
// working client struct.
func NewClient() (c *Client, err error) {
	c = new(Client)
	c.genericWallets = make(map[state.WalletID]Keypair)

	err = c.processConfigFile()
	if err != nil {
		return
	}

	// If there is an auto-connect feature, it should be specified in the
	// config file, and managed by processConfigFile

	// More stuff may go here.

	return
}
