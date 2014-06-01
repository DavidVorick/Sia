package participant

import (
	"fmt"
	"network"
	"quorum"
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
	heartbeats     [quorum.QuorumSize]map[siacrypto.TruncatedHash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock sync.Mutex

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

func (p *Participant) Subscribe(a network.Address, arb *struct{}) (err error) {
	// add the address to listeners
	p.listenersLock.Lock()
	p.listeners = append(p.listeners, a)
	p.listenersLock.Unlock()
	return
}

func (p *Participant) TransferQuorum(arb *struct{}, encodedQuorum *[]byte) (err error) {
	// lock the quorum before making major changes
	gobQuorum, err := p.quorum.GobEncode()
	if err != nil {
		return
	}
	*encodedQuorum = gobQuorum
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

// CreateParticipant creates a participant.
func CreateParticipant(messageRouter network.MessageRouter) (p *Participant, err error) {
	// check for non-nil messageRouter
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// create a signature keypair for this participant
	pubKey, secKey, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Can this be merged into one step?
	address := messageRouter.Address()
	address.ID = messageRouter.RegisterHandler(p)

	// initialize State with default values and keypair
	p = &Participant{
		messageRouter: messageRouter,
		self: quorum.NewSibling(address, pubKey),
		secretKey:   secKey,
		currentStep: 1,
	}

	// initialize heartbeat maps
	for i := range p.heartbeats {
		p.heartbeats[i] = make(map[siacrypto.TruncatedHash]*heartbeat)
	}

	// if we are the bootstrap participant, initialize a new quorum
	if p.self.Address() == bootstrapAddress {
		p.quorum.AddSibling(p.self)
		p.newSignedHeartbeat()
		go p.tick()
		return
	}

	// send a listener request to the bootstrap to become a listener on the quorum
	fmt.Println("Synchronizing to Bootstrap")
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.Subscribe",
		Args: p.self.Address(),
		Resp: nil,
	})
	if err != nil {
		return
	}

	// Get the current quorum struct
	q := new(quorum.Quorum)
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.TransferQuorum",
		Args: nil,
		Resp: q,
	})
	if err != nil {
		return
	}
	p.quorum = *q

	// Synchronize to the current quorum
	synchronize := new(Synchronize)
	err = p.messageRouter.SendMessage(&network.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.Synchronize",
		Args: nil,
		Resp: synchronize,
	})
	if err != nil {
		return
	}

	go p.tick()

	// join the network as a sibling

	return
}
