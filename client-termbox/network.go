package main

import (
	"github.com/NebulousLabs/Sia/network"
)

type NetworkState struct {
	Router *network.RPCServer

	ServerAddress network.Address
}

var networkState NetworkState
