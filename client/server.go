package client

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// The Server houses all of the participants. It will eventually contain a
// clock object that will be used and modified by all participants.
type Server struct {
	participants map[string]*consensus.Participant
}

// NewServer takes a port number as input and returns a server object that's
// ready to be populated with participants.
func (c *Client) NewServer() (err error) {
	// If the network router is nil, a server can't exist.
	if c.router == nil {
		err = errors.New("need to have a connection before creating a server")
		return
	}

	// Prevent any existing server from being overwritten.
	if c.participantServer != nil {
		err = errors.New("server already exists")
		return
	}

	// Establish c.participantServer.
	c.participantServer = new(Server)
	c.participantServer.participants = make(map[string]*consensus.Participant)
	return
}

// NewParticipant creates a directory 'name' at location 'filepath' and then
// creates a participant that will use that directory for its files. It's
// mostly a helper function to eliminate redundant code.
func (c *Client) createParticipantStructures(name string, filepath string) (fullname string, err error) {
	// NEED TO DO A CHECK ON IF THE NAME IS FILESYSTEM SAFE

	// Check that the participant server has been created.
	if c.participantServer == nil {
		err = errors.New("participant server is nil")
		return
	}

	// Check that a participant of the given name does not already exist.
	_, exists := c.participantServer.participants[name]
	if exists {
		err = errors.New("a participant of that name already exists.")
		return
	}

	// Create a directory 'name' at location 'filepath' for use of the
	// participant.
	fullname = path.Join(filepath, name) + "/"
	err = os.MkdirAll(fullname, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	return
}

// NewBootstrapParticipant creates a new participant that is the first in it's
// quorum; it creates the quorum along with the participant.
func (c *Client) NewBootstrapParticipant(name string, filepath string, sibID state.WalletID) (err error) {
	fullname, err := c.createParticipantStructures(name, filepath)
	if err != nil {
		return
	}

	// Create a keypair for the wallet that the sibling will tether to.
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Create the participant and add it to the server map.
	newParticipant, err := consensus.CreateBootstrapParticipant(c.router, fullname, sibID, pk)
	if err != nil {
		return
	}
	c.participantServer.participants[name] = newParticipant

	// Add the wallet to the client list of generic wallets.
	c.genericWallets[sibID] = Keypair{
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
	c.siblings = metadata.Siblings

	return
}

// NewJoiningParticipant creates a participant that joins the network known to
// the client as a sibling.
func (c *Client) NewJoiningParticipant(name string, filepath string, sibID state.WalletID) (err error) {
	fullname, err := c.createParticipantStructures(name, filepath)
	if err != nil {
		return
	}

	// Verify that the sibID given is available to the client.
	_, exists := c.genericWallets[sibID]
	if !exists {
		err = errors.New("no known wallet of that id")
		return
	}

	// Get a list of addresses for the joining participant to use while bootstrapping.
	var siblingAddresses []network.Address
	for _, sibling := range c.siblings {
		siblingAddresses = append(siblingAddresses, sibling.Address)
	}

	joiningParticipant, err := consensus.CreateJoiningParticipant(c.router, fullname, sibID, c.genericWallets[sibID].SecretKey, siblingAddresses)
	if err != nil {
		return
	}
	c.participantServer.participants[name] = joiningParticipant

	// Update the list of siblings to contain the bootstrap address, by
	// getting a list of siblings out of the joiningParticipant metadata.
	var metadata state.Metadata
	err = joiningParticipant.Metadata(struct{}{}, &metadata)
	if err != nil {
		return
	}
	c.siblings = metadata.Siblings

	return
}

// ParticipantMetadata returns the metadata for the participant with the given
// name. If no participant of that name exists, an error is returned.
func (c *Client) ParticipantMetadata(name string) (m state.Metadata, err error) {
	participant, exists := c.participantServer.participants[name]
	if !exists {
		err = errors.New("no participant of that name found")
		return
	}

	err = participant.Metadata(struct{}{}, &m)
	if err != nil {
		return
	}

	return
}

// ParticipantWallets returns every wallet known to the participant of the
// given name.
func (c *Client) ParticipantWallets(name string) (wallets []state.Wallet, err error) {
	participant, exists := c.participantServer.participants[name]
	if !exists {
		err = errors.New("no participant of that name found")
		return
	}

	var walletIDList []state.WalletID
	err = participant.WalletIDs(struct{}{}, &walletIDList)
	if err != nil {
		return
	}
	fmt.Println(walletIDList)

	for _, id := range walletIDList {
		var wallet state.Wallet
		err = participant.Wallet(id, &wallet)
		if err != nil {
			return
		}

		wallets = append(wallets, wallet)
	}
	fmt.Println(wallets)

	return
}
