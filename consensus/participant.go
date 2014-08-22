package consensus

import (
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
	address network.Address
	router  *network.RPCServer

	// Update Variables
	updates            [state.QuorumSize]map[siacrypto.Hash]Update
	scriptInputs       []state.ScriptInput
	updateAdvancements []state.UpdateAdvancement
	updatesLock        sync.RWMutex

	// Consensus Algorithm Status
	ticking     bool
	tickStart   time.Time
	currentStep byte
	tickLock    sync.RWMutex
	updateStop  sync.RWMutex
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
	for i, sibling := range p.engine.Metadata().Siblings {
		if !sibling.Inactive() && i != int(p.siblingIndex) {
			message.Dest = sibling.Address
			p.router.SendAsyncMessage(message)
		}
	}
	p.engineLock.Unlock()
}
