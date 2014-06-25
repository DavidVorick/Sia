package client

import (
	"network"
	"participant"
)

type Client struct {
	router *network.RPCServer
}

func (c *Client) Connect() (err error) {
	router, err := network.NewRPCServer(9989)
	if err != nil {
		return
	}
	err = router.Ping(&participant.BootstrapAddress)
	return
}
