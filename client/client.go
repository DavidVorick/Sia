package client

import (
	"errors"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// Keypair contains a public key and its corresponding private key. The keypair
// is given its own struct to enforce the connection between the keys.
type Keypair struct {
	PublicKey    siacrypto.PublicKey
	SecretKey    siacrypto.SecretKey
	OriginalSize int64
}

// Struct Client contains the state for client actions
type Client struct {
	// Networking Variables
	router   *network.RPCServer
	siblings [state.QuorumSize]state.Sibling

	// All Generic Wallets
	genericWallets map[state.WalletID]Keypair

	// Participant Server
	participantServer *Server
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

// IsRouterInitialized is useful for telling front end programs whether a
// router needs to be initialized or not.
func (c *Client) IsRouterInitialized() bool {
	return c.router != nil
}

// IsServerInitialized() is useful for telling front ent programs whether a
// server needs to be initialized or not.
func (c *Client) IsServerInitialized() bool {
	return c.participantServer != nil
}

// There should probably be some sort of error checking, but I'm not sure the best approach to that.
func (c *Client) Broadcast(m network.Message) {
	for i := range c.siblings {
		if c.siblings[i].Address.Host == "" {
			continue
		}
		m.Dest = c.siblings[i].Address
		c.router.SendMessage(m)
		break
	}
}

// Connect will create a new router, listening on the input port. If there is
// already a non-nil router in the client, an error will be returned.
func (c *Client) Connect(port uint16) (err error) {
	if c.router != nil {
		err = errors.New("network router has already been initialized")
		return
	}

	c.router, err = network.NewRPCServer(port)
	if err != nil {
		return
	}
	return
}

// Get height talks to sibling 0 and returns the height of the quorum, as
// reported by sibling 0.
func (c *Client) GetHeight() (height uint32, err error) {
	var m state.Metadata
	err = c.router.SendMessage(network.Message{
		Dest: c.siblings[0].Address,
		Proc: "Participant.Metadata",
		Args: struct{}{},
		Resp: &m,
	})
	if err != nil {
		return
	}
	height = m.Height
	return
}

// Figure out the latest list of siblings in the quorum.
func (c *Client) RefreshSiblings() (err error) {
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
			// Prevents all but one batch from getting through.
			break
		}
	}
	if err != nil {
		err = errors.New("Could not reach any stored siblings")
		return
	}

	// Right now the function just uses the first batch of siblings that
	// are recieved. In the future it will instead do some smart comparison
	// and pick the batch of siblings that seems most likely of all the
	// batches it receives.
	c.siblings = metadata.Siblings

	return
}

/*
// Closes and destroys the client's RPC server
func (c *Client) Disconnect() {
	if c.router == nil {
		return
	}
	c.router.Close()
	c.router = nil
}
*/

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
	c.siblings = metadata.Siblings
	return
}
