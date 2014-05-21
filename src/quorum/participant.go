package quorum

import (
	"common"
	"common/crypto"
	"fmt"
	"sync"
)

type Participant struct {
	// The quorum in which the participant participates
	quorum quorum

	// Variables local to the participant
	messageRouter common.MessageRouter
	self          *Sibling         // the sibling object for this participant
	secretKey     crypto.SecretKey // secret key matching self.publicKey

	// Heartbeat Variables
	currHeartbeat     heartbeat // in-progress heartbeat
	currHeartbeatLock sync.Mutex
	heartbeats        [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock    sync.Mutex

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

// CreateParticipant creates a participant.
func CreateParticipant(messageRouter common.MessageRouter) (p *Participant, err error) {
	// check for non-nil messageRouter
	if messageRouter == nil {
		err = fmt.Errorf("Cannot initialize with a nil messageRouter")
		return
	}

	// create a signature keypair for this participant
	pubKey, secKey, err := crypto.CreateKeyPair()
	if err != nil {
		return
	}

	// initialize State with default values and keypair
	p = &Participant{
		messageRouter: messageRouter,
		self: &Sibling{
			index:     255,
			address:   messageRouter.Address(),
			publicKey: pubKey,
		},
		secretKey:   secKey,
		currentStep: 1,
	}

	// register State and store our assigned ID
	p.self.address.ID = messageRouter.RegisterHandler(p)

	// if we are the bootstrap participant, initialize
	if p.self.address == bootstrapAddress {
		p.self.index = 0
		p.addNewSibling(p.self)
		p.ticking = true
		go p.tick()
		return
	}
	// otherwise, send a join request to the bootstrap
	fmt.Println("joining network...")
	errChan := p.messageRouter.SendAsyncMessage(&common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.JoinSia",
		Args: *p.self,
		Resp: nil,
	})
	err = <-errChan
	return
}

// Takes a Message and broadcasts it to every Sibling in the quorum
func (p *Participant) broadcast(m *common.Message) {
	p.quorum.siblingsLock.RLock()
	for i := range p.quorum.siblings {
		if p.quorum.siblings[i] != nil {
			nm := *m
			nm.Dest = p.quorum.siblings[i].address
			p.messageRouter.SendAsyncMessage(&nm)
		}
	}
	p.quorum.siblingsLock.RUnlock()
}
