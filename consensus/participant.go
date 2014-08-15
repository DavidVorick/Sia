package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// TODO: add docstring
type Participant struct {
	engine     delta.Engine
	engineLock sync.RWMutex

	// Variables local to the participant
	siblingIndex byte
	publicKey    siacrypto.PublicKey
	secretKey    siacrypto.SecretKey

	// Network Related Variables
	address       network.Address
	messageRouter network.MessageRouter

	// Update Variables
	updates            [state.QuorumSize]map[siacrypto.Hash]Update
	scriptInputs       []delta.ScriptInput
	updateAdvancements []state.UpdateAdvancement
	updatesLock        sync.RWMutex

	// Consensus Algorithm Status
	ticking     bool
	tickStart   time.Time
	currentStep byte
	tickLock    sync.RWMutex
}

var errNilMessageRouter = errors.New("cannot create a participant with a nil message router")

// NewParticipant initializes a Participant object with the provided
// MessageRouter and filePrefix. It also creates a keypair and sets default
// values for the siblingIndex and currentStep.
func newParticipant(mr network.MessageRouter, filePrefix string) (p *Participant, err error) {
	if mr == nil {
		err = errNilMessageRouter
		return
	}

	p = new(Participant)

	// Create a keypair for the participant.
	p.publicKey, p.secretKey, err = siacrypto.CreateKeyPair()
	if err != nil {
		return
	}
	p.siblingIndex = ^byte(0)
	p.currentStep = 1

	// Create the update maps.
	for i := range p.updates {
		p.updates[i] = make(map[siacrypto.Hash]Update)
	}

	// Initialize the network components of the participant.
	p.address = mr.RegisterHandler(p)
	p.messageRouter = mr

	// Initialize the file prefix
	p.engine.SetFilePrefix(filePrefix)
	p.engine.SetSiblingIndex(p.siblingIndex)

	return
}

// Ping is the simplest RPC possible. It exists only to confirm that a
// participant is reachable and listening. Ping should be called via
// RPCServer.Ping() instead of RPCServer.SendMessage().
func (p *Participant) Ping(_ struct{}, _ *struct{}) error {
	return nil
}

// broadcast sends a message to every sibling in the quorum. It cannot be used
// when the response value needs to be checked. It also discards any errors
// received.
func (p *Participant) broadcast(message network.Message) {
	// Send the message to all active and passive siblings in the quorum.
	p.engineLock.Lock()
	for _, sibling := range p.engine.Metadata().Siblings {
		if !sibling.Inactive() {
			message.Dest = sibling.Address
			p.messageRouter.SendAsyncMessage(message)
		}
	}
	p.engineLock.Unlock()
}
