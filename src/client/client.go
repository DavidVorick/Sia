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
	router, err := network.NewRPCServer(9989)
	if err != nil {
		return
	}
	err = router.Ping(&participant.BootstrapAddress)
	return
}

func (c *Client) Init() (err error) {
	c.genericWallets = make(map[quorum.WalletID]*siacrypto.Keypair)
	err = c.Connect()
	return
}
