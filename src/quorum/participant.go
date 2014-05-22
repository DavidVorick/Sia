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

// Update the state according to the information presented in the heartbeat
//
// What if a hopeful is denied because the quorum is full, but then later a
// participant gets tossed. This is really a question of when updates should
// be processed. Should they be processed before or after the participants
// are processed? Should proccessUpdates be its own funciton?
func (p *Participant) processHeartbeat(hb *heartbeat, seed *common.Entropy) (err error) {
	// add hopefuls to any available slots
	// quorum is already locked by compile()
	j := 0
	for _, s := range hb.hopefuls {
		for j < common.QuorumSize {
			if p.quorum.siblings[j] == nil {
				// transfer the quorum to the new sibling
				go func() {
					// wait until compile() releases the mutex
					p.quorum.lock.RLock()
					gobQuorum, _ := p.quorum.GobEncode()
					p.quorum.lock.RUnlock() // quorum can be unlocked as soon as GobEncode() completes
					p.messageRouter.SendAsyncMessage(&common.Message{
						Dest: s.address,
						Proc: "Participant.TransferQuorum",
						Args: gobQuorum,
						Resp: nil,
					})
				}()
				s.index = byte(j)
				p.addNewSibling(s)
				println("placed hopeful at index", j)
				break
			}
			j++
		}
	}

	// Add the entropy to newSeed
	th, err := crypto.CalculateTruncatedHash(append(seed[:], hb.entropy[:]...))
	copy(seed[:], th[:])

	return
}

// Takes a Message and broadcasts it to every Sibling in the quorum
// Even sends the message to self, this may be revised
func (p *Participant) broadcast(m *common.Message) {
	p.quorum.lock.RLock()
	for i := range p.quorum.siblings {
		if p.quorum.siblings[i] != nil {
			nm := *m
			nm.Dest = p.quorum.siblings[i].address
			p.messageRouter.SendAsyncMessage(&nm)
		}
	}
	p.quorum.lock.RUnlock()
}
