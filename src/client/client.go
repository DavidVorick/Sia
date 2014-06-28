package client

import (
	"network"
	"participant"
	"path/filepath"
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
	return
}

// This new function is a bit unique because it can return an error while also
// returning a fully working client.
func NewClient() (c *Client, err error) {
	c = new(Client)
	c.genericWallets = make(map[quorum.WalletID]*siacrypto.Keypair)
	filenames, err := filepath.Glob("*.id")
	if err != nil {
		panic(err)
	}
	for _, j := range filenames {
		id, keypair, err := LoadWallet(j)
		if err != nil {
			panic(err)
		}
		c.genericWallets[id] = keypair
	}
	err = c.Connect("localhost", 9988) // default bootstrap address
	return
}
