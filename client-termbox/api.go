package main

import (
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

func serverAddress() network.Address {
	return network.Address{
		config.Server.Hostname,
		config.Server.Port,
		network.Identifier(config.Server.ID),
	}
}

// Fetch a list of wallets from the server.
func getWallets() (ids []state.WalletID, err error) {
	err = config.Router.SendMessage(network.Message{
		Dest: serverAddress(),
		Proc: "Server.WalletIDs",
		Args: struct{}{},
		Resp: &ids,
	})
	return
}
