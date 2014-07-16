package consensus

/* import (
	"bytes"
	"delta"
	"encoding/gob"
	"quorum"
	"siacrypto"
)

type SnapshotWalletsInput struct {
	Snapshot bool
	Ids      []quorum.WalletID
}

// Contains Synchronization information for the quorum.
// Eventually this should include an offset.
type Synchronize struct {
	currentStep int
	heartbeats  [quorum.QuorumSize]map[siacrypto.Hash]*heartbeat
}

func (p *Participant) RecentSnapshot(_ struct{}, q *quorum.Quorum) (err error) {
	//quorum, err := p.quorum.RecentSnapshot()
	//*q = *quorum
	return
}

func (p *Participant) SnapshotWalletList(snapshot bool, ids *[]quorum.WalletID) (err error) {
	//*ids = p.quorum.SnapshotWalletList(snapshot)
	return
}

func (p *Participant) SnapshotWallets(swi SnapshotWalletsInput, wallets *[][]byte) (err error) {
	//*wallets = p.quorum.SnapshotWallets(swi.Snapshot, swi.Ids)
	return
}

func (p *Participant) SnapshotBlocks(snapshot bool, blockList *[]delta.Block) (err error) {
	//*blockList = p.loadBlocks(snapshot)
	return
}

// Participant.Siblings is an RPC call that returns a set of quorum.QuorumSize
// siblings that have been encoded individually into a byte slice. This is
// necessary because gob doesn't know how to understand slices or arrays of
// structs (somewhat frustratingly).
func (p *Participant) Siblings(_ struct{}, encodedSiblings *[]byte) (err error) {
	siblings := p.quorum.Siblings()
	gobSiblings, err := quorum.EncodeSiblings(siblings)
	if err != nil {
		return
	}

	*encodedSiblings = gobSiblings
	return
}

func (p *Participant) Synchronize(_ struct{}, s *Synchronize) (err error) {
	p.stepLock.Lock()
	s.currentStep = p.currentStep
	p.stepLock.Unlock()

	p.heartbeatsLock.Lock()
	s.heartbeats = p.heartbeats
	p.heartbeatsLock.Unlock()
	return
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
} */
