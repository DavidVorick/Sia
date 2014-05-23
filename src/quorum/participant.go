package quorum

import (
	"common"
	"common/crypto"
	"fmt"
	"sync"
)

type Update interface {
	process(p *Participant)

	// parent block

	GobEncode() ([]byte, error)
	GobDecode([]byte) error
}

type Participant struct {
	// The quorum in which the participant participates
	quorum quorum

	// Variables local to the participant
	self      *Sibling         // the sibling object for this participant
	secretKey crypto.SecretKey // secret key matching self.publicKey

	// Network Related Variables
	messageRouter common.MessageRouter
	listeners     []common.Address

	// Heartbeat Variables
	updates        map[Update]*Update
	updatesLock    sync.Mutex
	heartbeats     [common.QuorumSize]map[crypto.TruncatedHash]*heartbeat // list of heartbeats received from siblings
	heartbeatsLock sync.Mutex

	// Consensus Algorithm Status
	currentStep int
	stepLock    sync.RWMutex // prevents a benign race condition
	ticking     bool
	tickingLock sync.Mutex
}

// AddUpdate takes an update from an arbitrary source and includes it in the
// next heartbeat, kind of like a miner in Bitcoin will queue a transaction to
// be added in the next block.
func (p *Participant) AddUpdate(u Update, arb *struct{}) (err error) {
	// to be added: check the update for being valid, as to not waste bandwidth
	p.updatesLock.Lock()
	p.updates[u] = &u
	p.updatesLock.Unlock()
	return
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

	// if we are the bootstrap participant, initialize a new quorum
	if p.self.address == bootstrapAddress {
		p.self.index = 0
		p.heartbeats[p.self.index] = make(map[crypto.TruncatedHash]*heartbeat)
		p.quorum.siblings[p.self.index] = p.self
		go p.tick()
		return
	}

	// if we are not the bootstrap, send a join request to the bootstrap
	// first create a join Update

	j := Join{
		Sibling: *p.self,
	}

	// j.GobEncode ===> make it an encoded update
	// then send the encoded update over RPC to an Update receiver, instead of a join receiver

	fmt.Println("joining network...")
	errChan := p.messageRouter.SendAsyncMessage(&common.Message{
		Dest: bootstrapAddress,
		Proc: "Participant.JoinSia",
		Args: j,
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
func (p *Participant) processHeartbeat(hb *heartbeat, seed *common.Entropy, updateList map[Update]bool) (err error) {
	// Add the entropy to newSeed
	th, err := crypto.CalculateTruncatedHash(append(seed[:], hb.entropy[:]...))
	copy(seed[:], th[:])

	// Process updates and add to update list
	for _, update := range hb.updates {
		if updateList[*update] == false {
			(*update).process(p)
			updateList[*update] = true
		}
	}

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

// Right now there's no way to stop listening. Eventually, listeners will also
// have public keys which they can use to signal that they wish to stop
// listening. We can probably also add timeouts so that listeners are
// automatically ignored after M days andmust be renewed.
func (p *Participant) QuorumListenerRequest(a common.Address, arb *struct{}) (err error) {
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
	p.messageRouter.SendAsyncMessage(&common.Message{
		Dest: a,
		Proc: "Participant.TransferQuorum",
		Args: gobQuorum,
		Resp: nil,
	})
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

	fmt.Println("downloaded quorum:")
	fmt.Print(p.quorum.Status())

	// create maps for each sibling
	p.quorum.lock.RLock()
	for i := range p.quorum.siblings {
		p.heartbeats[i] = make(map[crypto.TruncatedHash]*heartbeat)
	}
	p.quorum.lock.RUnlock()
	go p.tick()
	return
}
