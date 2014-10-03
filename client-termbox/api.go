package main

import (
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

func (s *Server) CreateParticipant(name string, id uint64, dir string, bootstrap bool) error {
	// construct NewParticipantInfo
	// this type is defined in the server's main package, so we have to
	// recreate it here.
	npi := struct {
		Name               string
		SiblingID          state.WalletID
		UseUniqueDirectory bool
		UniqueDirectory    string
	}{
		Name:            name,
		SiblingID:       state.WalletID(id),
		UniqueDirectory: dir,
	}
	if dir != "" {
		npi.UseUniqueDirectory = true
	}

	var proc string
	if bootstrap {
		proc = "NewBootstrapParticipant"
	} else {
		proc = "NewJoiningParticipant"
	}

	return s.Router.SendMessage(network.Message{
		Dest: s.Address,
		Proc: "Server." + proc,
		Args: npi,
	})
}

func (s *Server) FetchMetadata(name string, metadata *state.Metadata) error {
	return s.Router.SendMessage(network.Message{
		Dest: s.Address,
		Proc: "Server.ParticipantMetadata",
		Args: name,
		Resp: metadata,
	})
}
