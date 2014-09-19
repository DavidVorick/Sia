package main

import (
	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"
)

type Server struct {
	Address network.Address
	Router  *network.RPCServer
}

func (s *Server) UpdateAddress() {
	s.Address = network.Address{
		config.Server.Host,
		config.Server.Port,
		network.Identifier(config.Server.ID),
	}
}

func (s *Server) GetWallets() (ids []state.WalletID, err error) {
	err = s.Router.SendMessage(network.Message{
		Dest: s.Address,
		Proc: "Server.WalletIDs",
		Args: struct{}{},
		Resp: &ids,
	})
	return
}

func (s *Server) GetParticipantNames() (names []string, err error) {
	err = s.Router.SendMessage(network.Message{
		Dest: s.Address,
		Proc: "Server.ParticipantNames",
		Args: struct{}{},
		Resp: &names,
	})
	return
}
