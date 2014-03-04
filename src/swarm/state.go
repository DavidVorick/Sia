package swarm

import (
	"common"
	"sync"
)

//List of states
// SwarmInformed - Swarm member shave been told to join swarm
// SwarmConnected - Swarm Members have succesfully formed a block
// SwarmLive - Swarm has sent a signal to the parent blockchain saying it is
//             alive and is in the steady state
// SwarmJoin - We are joining an already alive swarm
// SwarmDied - The swarm has died, terminate
type State interface {
	HandleUpdate(t common.Update) State
}

func newBlockchain(Host string, Id string, StorageMapping map[string]interface{}) (b *Blockchain) {
	b = new(Blockchain)
	b.Host = Host
	b.Id = Id
	b.StorageMapping = StorageMapping
	b.outgoingUpdates = make(chan common.Update)
	b.incomingMessages = make(chan common.NetworkMessage)
	b.SeenTransactions = make(map[string]bool)
	b.lock = &sync.Mutex{}
	return
}

func NewBlockchain(Host string, Id string, StorageMapping map[string]interface{}) (b *Blockchain) {
	b = newBlockchain(Host, Id, StorageMapping)
	b.state = NewStateSwarmInformed(b)
	return
}
