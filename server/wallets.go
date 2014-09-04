package server

import (
	"errors"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// Returns a list of all wallets available to the client.
func (s *Server) GetWalletIDs() (ids []state.WalletID) {
	ids = make([]state.WalletID, 0, len(s.genericWallets))
	for id := range s.genericWallets {
		ids = append(ids, state.WalletID(id))
	}
	return
}

// Wallet type takes an id as input and returns the wallet type. An error is
// returned if the wallet is not found by the client.
func (s *Server) WalletType(id state.WalletID) (walletType string, err error) {
	// Check if the wallet is a generic type.
	_, exists := s.genericWallets[GenericWalletID(id)]
	if exists {
		walletType = "generic"
		return
	}

	// Check for other types of wallets.

	err = errors.New("Wallet is not available.")
	return
}

// Submit a wallet request to the fountain wallet.
func (s *Server) RequestGenericWallet(id state.WalletID) (err error) {
	// Query to verify that the wallet id is available.
	var w state.Wallet
	err = s.router.SendMessage(network.Message{
		Dest: s.metadata.Siblings[0].Address,
		Proc: "Participant.Wallet",
		Args: id,
		Resp: &w,
	})
	if err == nil {
		err = errors.New("Wallet already exists!")
		return
	}
	err = nil

	// Create a generic wallet with a keypair for the request.
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Get the current height of the quorum.
	// Send the requesting script input out to the network.
	s.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: state.ScriptInput{
			WalletID: delta.FountainWalletID,
			Input:    delta.CreateFountainWalletInput(id, delta.DefaultScript(pk)),
			Deadline: s.metadata.Height + state.MaxDeadline,
		},
	})

	// Wait an appropriate amount of time for the request to be accepted: 2
	// blocks.
	time.Sleep(time.Duration(consensus.NumSteps) * 2 * consensus.StepDuration)

	// Query to verify that the request was accepted by the network.
	err = s.router.SendMessage(network.Message{
		Dest: s.metadata.Siblings[0].Address,
		Proc: "Participant.Wallet",
		Args: id,
		Resp: &w,
	})
	if err != nil {
		return
	}
	if string(w.Script) != string(delta.DefaultScript(pk)) {
		err = errors.New("Wallet already exists - someone just beat you to it.")
		return
	}

	// Fill out a keypair object and insert it into the generic wallet map.
	gw := GenericWallet{
		WalletID:  id,
		PublicKey: pk,
		SecretKey: sk,
	}
	s.genericWallets[GenericWalletID(id)] = &gw

	return
}
