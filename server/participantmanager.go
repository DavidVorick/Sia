package main

import (
	"errors"
	"os"
	"path"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// The ParticipantManager houses all of the participants. It will eventually contain a
// clock object that will be used and modified by all participants.
type ParticipantManager struct {
	parentDirectory string
	participants    map[string]*consensus.Participant
}

type NewParticipantInfo struct {
	Name      string
	SiblingID state.WalletID

	UseUniqueDirectory bool
	UniqueDirectory    string
}

// NewParticipantManager takes a port number as input and returns a server object that's
// ready to be populated with participants.
func newParticipantManager() (p *ParticipantManager, err error) {
	// Determine whether the server is public.
	if !config.Network.PublicConnection {
		err = errors.New("server is not public - the whole network will be local")
	}

	// Create the participant manager.
	p = new(ParticipantManager)
	p.participants = make(map[string]*consensus.Participant)
	return
}

// NewParticipant creates a directory 'name' at location 'filepath' and then
// creates a participant that will use that directory for its files. It's
// mostly a helper function to eliminate redundant code.
func (s *Server) createParticipantStructures(npi NewParticipantInfo) (dirname string, err error) {
	// Check that a participant of the given name does not already exist.
	_, exists := s.participantManager.participants[npi.Name]
	if exists {
		err = errors.New("a participant of that name already exists.")
		return
	}

	// Create a directory for the participant, using the unique directory
	// if the flag is set, and using the participantDir/Name if not.
	if npi.UseUniqueDirectory {
		dirname = npi.UniqueDirectory
	} else {
		dirname = path.Join(config.Filesystem.ParticipantDir, npi.Name) + "/"
	}
	err = os.MkdirAll(dirname, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	return
}

// NewBootstrapParticipant creates a new participant that is the first in it's
// quorum; it creates the quorum along with the participant.
func (s *Server) NewBootstrapParticipant(npi NewParticipantInfo, _ *struct{}) (err error) {
	dirname, err := s.createParticipantStructures(npi)
	if err != nil {
		return
	}

	// Create a keypair for the wallet that the sibling will tether to.
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Create the participant and add it to the server map.
	newParticipant, err := consensus.CreateBootstrapParticipant(s.router, dirname, npi.SiblingID, pk)
	if err != nil {
		return
	}
	s.participantManager.participants[npi.Name] = newParticipant

	// Add the wallet to the client list of generic wallets.
	s.genericWallets[GenericWalletID(npi.SiblingID)] = &GenericWallet{
		WalletID:  npi.SiblingID,
		PublicKey: pk,
		SecretKey: sk,
	}

	// Update the list of siblings to contain the bootstrap address, by
	// getting a list of siblings out of the newParticipant metadata.
	var metadata state.Metadata
	err = newParticipant.Metadata(struct{}{}, &metadata)
	if err != nil {
		return
	}
	s.metadata.Siblings = metadata.Siblings

	return
}

// NewJoiningParticipant creates a participant that joins the network known to
// the client as a sibling.
func (s *Server) NewJoiningParticipant(npi NewParticipantInfo, _ *struct{}) (err error) {
	dirname, err := s.createParticipantStructures(npi)
	if err != nil {
		return
	}

	// Verify that the sibID given is available to the client.
	_, exists := s.genericWallets[GenericWalletID(npi.SiblingID)]
	if !exists {
		err = errors.New("no known wallet of that id")
		return
	}

	// Refresh the metadata and then grab all the addresses for the joining
	// participant to use.
	err = s.refreshMetadata()
	if err != nil {
		return
	}
	var siblingAddresses []network.Address
	for _, sibling := range s.metadata.Siblings {
		siblingAddresses = append(siblingAddresses, sibling.Address)
	}

	joiningParticipant, err := consensus.CreateJoiningParticipant(s.router, dirname, npi.SiblingID, s.genericWallets[GenericWalletID(npi.SiblingID)].SecretKey, siblingAddresses)
	if err != nil {
		return
	}
	s.participantManager.participants[npi.Name] = joiningParticipant

	// Update the list of siblings to contain the bootstrap address, by
	// getting a list of siblings out of the joiningParticipant metadata.
	var metadata state.Metadata
	err = joiningParticipant.Metadata(struct{}{}, &metadata)
	if err != nil {
		return
	}
	s.metadata.Siblings = metadata.Siblings

	return
}

// ParticipantMetadata returns the metadata for the participant with the given
// name. If no participant of that name exists, an error is returned.
func (s *Server) ParticipantMetadata(name string, m *state.Metadata) (err error) {
	participant, exists := s.participantManager.participants[name]
	if !exists {
		err = errors.New("no participant of that name found")
		return
	}

	err = participant.Metadata(struct{}{}, m)
	if err != nil {
		return
	}

	return
}

// ParticipantWallets returns every wallet known to the participant of the
// given name.
func (s *Server) ParticipantWallets(name string, wallets *[]state.Wallet) (err error) {
	participant, exists := s.participantManager.participants[name]
	if !exists {
		err = errors.New("no participant of that name found")
		return
	}

	var walletIDList []state.WalletID
	err = participant.WalletIDs(struct{}{}, &walletIDList)
	if err != nil {
		return
	}

	for _, id := range walletIDList {
		var wallet state.Wallet
		err = participant.Wallet(id, &wallet)
		if err != nil {
			return
		}

		*wallets = append(*wallets, wallet)
	}

	return
}
