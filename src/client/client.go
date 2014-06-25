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
func (c *Client) Connect() (err error) {
	c.router, err = network.NewRPCServer(9989)
	if err != nil {
		return
	}
	err = c.router.Ping(&participant.BootstrapAddress)
	return
}

// This new function is a bit unique because it can return an error while also
// returning a fully working client.
func NewClient() (c *Client, err error) {
	c = new(Client)
	c.genericWallets = make(map[quorum.WalletID]*siacrypto.Keypair)
	err = c.Connect()
	return
}
