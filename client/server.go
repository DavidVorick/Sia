package client

import (
	"errors"
	"os"
	"path"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/state"
)

// The Server houses all of the participants. It contains a single message
// router that is shared by all of the participants, it will eventually contain
// a clock object that will be used and modified by all participants.
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
// creates a participant that will use that directory for its files.
func (c *Client) NewParticipant(name string, filepath string, sibID state.WalletID) (err error) {
	// Check that a participant of the given name does not already exist.
	_, exists := c.participantServer.participants[name]
	if exists {
		err = errors.New("a participant of that name already exists.")
		return
	}

	// Create a directory 'name' at location 'filepath' for use of the
	// participant.
	fullname := path.Join(filepath, name) + "/"
	err = os.MkdirAll(fullname, os.ModeDir|os.ModePerm)
	if err != nil {
		return
	}

	// Create the participant and add it to the server map.
	newParticipant, err := consensus.CreateBootstrapParticipant(c.router, fullname, sibID)
	if err != nil {
		return
	}
	c.participantServer.participants[name] = newParticipant
	return
}

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
