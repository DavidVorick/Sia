package consensus

import (
	"delta"
	"errors"
	"network"
	"siacrypto"
	"state"
	"state/script"
	"sync"
)

type Participant struct {
	engine delta.Engine

	// Variables local to the participant
	self      *state.Sibling       // the sibling object for this participant
	secretKey *siacrypto.SecretKey // secret key matching self.publicKey

	// Network Related Variables
	messageRouter network.MessageRouter
	listeners     []network.Address
	listenersLock sync.RWMutex

	// Heartbeat Variables
	scriptInputs     []script.ScriptInput
	scriptInputsLock sync.Mutex
	//uploadAdvancements     []quorum.UploadAdvancement
	uploadAdvancementsLock sync.Mutex
	//heartbeats             [state.QuorumSize]map[siacrypto.Hash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock sync.Mutex

	// Consensus Algorithm Status
	ticking     bool
	tickingLock sync.Mutex
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition

	// Bootstrap variables
	synchronized bool
	recentBlocks map[uint32]*delta.Block

	// Block history variables
	activeHistoryStep int
	activeHistory     string // file currently being appended with new blocks
	recentHistory     string // file containing SnapshotLen blocks
}

var npNilMessageRouter = errors.New("Cannot create a participant with a nil message router.")

func NewParticipant(mr network.MessageRouter, filePrefix string) (p *Participant, err error) {
	if mr == nil {
		err = npNilMessageRouter
		return
	}

	p = new(Participant)

	// Create a keypair for the participant.
	publicKey, secretKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}
	p.secretKey = secretKey

	// Initialize the network components of the participant.
	p.messageRouter = mr
	p.self = &state.Sibling{
		Address:   mr.Address(),
		PublicKey: publicKey,
	}
	p.self.Address.ID = mr.RegisterHandler(p)

	// Initialize the file prefix
	p.engine.SetFilePrefix(filePrefix)

	// Initialize currentStep to 1
	p.currentStep = 1

	return
}

/* func (p *Participant) AddScriptInput(si script.ScriptInput, _ *struct{}) (err error) {
	p.scriptInputsLock.Lock()
	p.scriptInputs = append(p.scriptInputs, si)
	p.scriptInputsLock.Unlock()
	return
}

// Takes an address as input and adds the address to the list of listeners,
// meaning that the added address will get sent all messages that are broadcast
// to the quorum.
func (p *Participant) Subscribe(a network.Address, _ *struct{}) (err error) {
	// add the address to listeners
	p.listenersLock.Lock()
	p.listeners = append(p.listeners, a)
	p.listenersLock.Unlock()
	return
}

// Takes a message and broadcasts it to every sibling in the quorum and every
// listener subscribed to the participant
func (p *Participant) broadcast(m *network.Message) {
	// send the messagea to all of the siblings in the quorum
	siblings := p.quorum.Siblings()
	for _, sibling := range siblings {
		if sibling != nil {
			nm := *m
			nm.Dest = sibling.Address()
			p.messageRouter.SendAsyncMessage(&nm)
		}
	}

	for _, listener := range p.listeners {
		nm := *m
		nm.Dest = listener
		p.messageRouter.SendAsyncMessage(&nm)
	}
} */
