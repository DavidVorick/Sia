package quorum

import (
	"bytes"
	"common"
	"encoding/gob"
	"fmt"
	"siacrypto"
	"sync"
)

type Update interface {
	process(p *Participant)

	// parent block
}

type Synchronize struct {
	currentStep int
	heartbeats  [common.QuorumSize]map[siacrypto.TruncatedHash]*heartbeat
}

type Participant struct {
	// The quorum in which the participant participates
	quorum quorum

	// Variables local to the participant
	self      *Sibling            // the sibling object for this participant
	secretKey siacrypto.SecretKey // secret key matching self.publicKey

	// Network Related Variables
	messageRouter common.MessageRouter
	listeners     []common.Address

	// Heartbeat Variables
	updates        map[Update]Update
	updatesLock    sync.Mutex
	heartbeats     [common.QuorumSize]map[siacrypto.TruncatedHash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock sync.Mutex

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

func (s *Synchronize) GobEncode() (gobSynchronize []byte, err error) {
	if s == nil {
		s = new(Synchronize)
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err = encoder.Encode(s.currentStep)
	if err != nil {
		return
	}
	err = encoder.Encode(s.heartbeats)
	if err != nil {
		return
	}
	gobSynchronize = w.Bytes()
	return
}

func (s *Synchronize) GobDecode(gobSynchronize []byte) (err error) {
	if s == nil {
		s = new(Synchronize)
	}

	r := bytes.NewBuffer(gobSynchronize)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&s.currentStep)
	if err != nil {
		return
	}
	err = decoder.Decode(&s.heartbeats)
	if err != nil {
		return
	}
	return
}

// AddUpdate takes an update from an arbitrary source and includes it in the
// next heartbeat, kind of like a miner in Bitcoin will queue a transaction to
// be added in the next block.
func (p *Participant) AddUpdate(update Update, arb *struct{}) (err error) {
	// to be added: check the update for being valid, as to not waste bandwidth
	p.updatesLock.Lock()
	p.updates[update] = update
	p.updatesLock.Unlock()
	return
}

func (p *Participant) RequestSiblings(arb struct{}, siblings *[]*Sibling) (err error) {
	*siblings = p.quorum.siblings[:]
	return
}

// A member of a quorum will call TransferQuorum on someone who has solicited
// a quorum transfer. The quorum data is then sent between machines over RPC.
func (p *Participant) TransferQuorum(encodedQuorum []byte, arb *struct{}) (err error) {
	// lock the quorum before making major changes
	p.quorum.lock.Lock()
	err = p.quorum.GobDecode(encodedQuorum)
	p.quorum.lock.Unlock()
	if err != nil {
		return
	}

	// fmt.Println("downloaded quorum:")
	// fmt.Print(p.quorum.Status())

	// create maps for each sibling
	p.quorum.lock.RLock()
	p.heartbeatsLock.Lock()
	for i := range p.quorum.siblings {
		p.heartbeats[i] = make(map[siacrypto.TruncatedHash]*heartbeat)
	}
	p.heartbeatsLock.Unlock()
	p.quorum.lock.RUnlock()
	go p.tick()
	return
}

// Synchronize just sends over participant.currentStep, and is a temporary
// function. It's not secure or trusted and is highely exploitable.
func (p *Participant) Synchronize(s Synchronize, arb *struct{}) (err error) {
	p.stepLock.Lock()
	p.currentStep = s.currentStep
	p.stepLock.Unlock()

	p.heartbeatsLock.Lock()
	p.heartbeats = s.heartbeats
	p.heartbeatsLock.Unlock()
	return
}

// Right now there's no way to stop listening. Eventually, listeners will also
// have public keys which they can use to signal that they wish to stop
// listening. We can probably also add timeouts so that listeners are
// automatically ignored after M days andmust be renewed.
func (p *Participant) AddListener(a common.Address, arb *struct{}) (err error) {
	// add the address to listeners
	p.listeners = append(p.listeners, a)

	// transfer the quorum to the new listener
	// this will also be moved to a completely seperate function in the future
	p.quorum.lock.RLock()
	gobQuorum, err := p.quorum.GobEncode()
	if err != nil {
		// error logging?
		return
	}
	p.quorum.lock.RUnlock() // quorum unlocked after GobEncode() completes
	p.messageRouter.SendMessage(&common.Message{
		Dest: a,
		Proc: "Participant.TransferQuorum",
		Args: gobQuorum,
		Resp: nil,
	})

	// may need to check for an error on the send here

	// synchronize the other quorum to the current step (after stepping) and send
	// the current list of heartbeats

	// create the synchronize object
	var sync Synchronize
	p.stepLock.Lock()
	p.heartbeatsLock.Lock()
	sync.currentStep = p.currentStep
	sync.heartbeats = p.heartbeats
	p.stepLock.Unlock()
	p.heartbeatsLock.Unlock()

	// send the synnchronize object in a message
	p.messageRouter.SendMessage(&common.Message{
		Dest: a,
		Proc: "Participant.Synchronize",
		Args: sync,
		Resp: nil,
	})
	return
}

// Update the state according to the information presented in the heartbeat
//
// What if a hopeful is denied because the quorum is full, but then later a
// participant gets tossed. This is really a question of when updates should
// be processed. Should they be processed before or after the participants
// are processed? Should proccessUpdates be its own funciton?
func (p *Participant) processHeartbeat(hb *heartbeat, seed *common.Entropy, updateList map[Update]bool) (err error) {
	// Add the entropy to newSeed
	th, err := siacrypto.CalculateTruncatedHash(append(seed[:], hb.entropy[:]...))
	copy(seed[:], th[:])

	// Process updates and add to update list
	for _, update := range hb.updates {
		if updateList[update] == false {
			update.process(p)
			updateList[update] = true
		}
	}

	return
}

// Takes a Message and broadcasts it to every Sibling in the quorum
// Even sends the message to self, this may be revised
// After sending to all siblings, all listeners are also sent the message
func (p *Participant) broadcast(m *common.Message) {
	// send the messagea to all of the siblings in the quorum
	p.quorum.lock.RLock()
	for i := range p.quorum.siblings {
		if p.quorum.siblings[i] != nil {
			nm := *m
			nm.Dest = p.quorum.siblings[i].address
			p.messageRouter.SendAsyncMessage(&nm)
		}
	}
	p.quorum.lock.RUnlock()

	for _, listener := range p.listeners {
		nm := *m
		nm.Dest = listener
		p.messageRouter.SendAsyncMessage(&nm)
	}
}

// CreateParticipant creates a participant.
func CreateParticipant(messageRouter common.MessageRouter) (p *Participant, err error) {
	// Initializations
	gob.Register(JoinRequest{})

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
	// to-do: write a test for RegisterHandler related functions
	p.self.address.ID = messageRouter.RegisterHandler(p)

	// if we are the bootstrap participant, initialize a new quorum
	if p.self.address == bootstrapAddress {
		p.self.index = 0
		p.heartbeats[p.self.index] = make(map[siacrypto.TruncatedHash]*heartbeat)
		p.quorum.siblings[p.self.index] = p.self
		p.newSignedHeartbeat()
		go p.tick()
		return
	}

	// send a listener request to the bootstrap to become a listener on the quorum
	fmt.Println("Synchronizing to Bootstrap")
	err = p.messageRouter.SendMessage(&common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.AddListener",
		Args: p.self.address,
		Resp: nil,
	})
	if err != nil {
		return
	}

	// Create a Join update and submit it to the quorum
	var j Update = JoinRequest{
		Sibling: *p.self,
	}
	fmt.Println("joining network...")
	err = p.messageRouter.SendMessage(&common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.AddUpdate",
		Args: &j,
		Resp: nil,
	})
	return
}
