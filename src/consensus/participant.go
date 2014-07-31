package consensus

import (
	"delta"
	"errors"
	"network"
	"siacrypto"
	"state"
	"sync"
)

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
	updates                [state.QuorumSize]map[siacrypto.Hash]Update
	updatesLock            sync.Mutex
	scriptInputs           []delta.ScriptInput
	scriptInputsLock       sync.Mutex
	updateAdvancements     []state.UpdateAdvancement
	updateAdvancementsLock sync.Mutex

	// Consensus Algorithm Status
	//ticking     bool
	//tickingLock sync.Mutex
	currentStep     byte
	currentStepLock sync.RWMutex // prevents a benign race condition

	// Bootstrap variables
	synchronized bool
	//recentBlocks map[uint32]*delta.Block
}

var nperrNilMessageRouter = errors.New("Cannot create a participant with a nil message router.")

func NewParticipant(mr network.MessageRouter, filePrefix string) (p *Participant, err error) {
	if mr == nil {
		err = nperrNilMessageRouter
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

	// Initialize the network components of the participant.
	p.address = mr.Address()
	p.address.ID = mr.RegisterHandler(p)
	p.messageRouter = mr

	// Initialize the file prefix
	p.engine.Initialize(filePrefix, p.siblingIndex)

	return
}

// Dummy function to test connectivity
func (p *Participant) Ping(_ struct{}, _ *struct{}) error {
	return nil
}

// Sends a message to every sibling in the quorum.
func (p *Participant) broadcast(message network.Message) {
	// send the message to all of the siblings in the quorum
	for _, sibling := range p.engine.Metadata().Siblings {
		if sibling.Active {
			message.Dest = sibling.Address
			p.messageRouter.SendAsyncMessage(message)
		}
	}
}

func (p *Participant) AddScriptInput(si delta.ScriptInput, _ *struct{}) (err error) {
	p.scriptInputsLock.Lock()
	p.scriptInputs = append(p.scriptInputs, si)
	p.scriptInputsLock.Unlock()
	return
}
