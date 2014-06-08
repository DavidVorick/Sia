package participant

import (
	"network"
	"quorum"
	"quorum/script"
	"siacrypto"
	"sync"
)

type Participant struct {
	// The quorum in which the participant participates
	quorum quorum.Quorum

	// Variables local to the participant
	self      *quorum.Sibling      // the sibling object for this participant
	secretKey *siacrypto.SecretKey // secret key matching self.publicKey

	// Network Related Variables
	messageRouter network.MessageRouter
	listeners     []network.Address
	listenersLock sync.RWMutex

	// Heartbeat Variables
	scriptInputs     []script.ScriptInput
	scriptInputsLock sync.Mutex
	heartbeats       [quorum.QuorumSize]map[siacrypto.TruncatedHash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock   sync.Mutex

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

func (p *Participant) AddScriptInput(si script.ScriptInput, _ *struct{}) (err error) {
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
}
