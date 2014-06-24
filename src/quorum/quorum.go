// A List of Calls Available To Script:
// 1. Send
// 2. AddNewSibling
// 3. CreateWallet
package quorum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"siacrypto"
	"sync"
)

const (
	QuorumSize     int = 4        // max siblings per quorum
	AtomSize       int = 4096     // in bytes
	AtomsPerQuorum int = 16777216 // 64GB
)

// A quorum is a set of data that is identical across all participants in the
// quorum. Data in the quorum can only be updated during a block, and the
// update must be deterministic and reversable.
type Quorum struct {
	// Quorum-wide lock
	lock sync.RWMutex

	// Network Variables
	siblings [QuorumSize]*Sibling

	// Compile Variables
	germ Entropy // Where internal entropy is stored before external entropy is applied
	seed Entropy // Used to generate random numbers during compilation

	// Event management
	eventCounter uint32
	eventRoot    *eventNode

	// Wallet management
	walletPrefix string
	wallets      uint32
	walletRoot   *walletNode

	// File management
	storagePrice uint32
	uploads      map[WalletID][]*upload

	// Snapshot management
	currentSnapshot bool // false == snap0, true == snap1

	// Block tracking
	parent siacrypto.Hash
	height uint32
}

// This is the prefix that the quorum will use when opening wallets as files.
// Eventually, logic will be implemented to move all of the wallets and files
// if the prefex is changed.
func (q *Quorum) SetWalletPrefix(walletPrefix string) {
	q.walletPrefix = walletPrefix
}

func (q *Quorum) GetWalletPrefix() string {
	return q.walletPrefix
}

func (q *Quorum) CurrentSnapshot() bool {
	return q.currentSnapshot
}

// Getter for the siblings private variable
func (q *Quorum) Siblings() [QuorumSize]*Sibling {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.siblings
}

func (q *Quorum) Height() uint32 {
	return q.height
}

func (q *Quorum) Parent() siacrypto.Hash {
	return q.parent
}

func (q *Quorum) AdvanceBlock(h siacrypto.Hash) {
	q.parent = h
	q.height += 1
}

// q.Status() enumerates the variables of the quorum in a human-readable output
func (q *Quorum) Status() (b string) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	b = "\nQuorum Status:\n"

	b += fmt.Sprintf("\tPrefix: %v\n\n", q.walletPrefix)
	b += fmt.Sprintf("\tSiblings:\n")
	for _, s := range q.siblings {
		if s != nil {
			pubKeyHash := s.publicKey.Hash()
			b += fmt.Sprintf("\t\t%v\n", s.index)
			b += fmt.Sprintf("\t\t\tAddress: %v\n", s.address)
			b += fmt.Sprintf("\t\t\tPublic Key: %v\n", pubKeyHash[:6])
		}
	}
	b += fmt.Sprintf("\n")

	b += fmt.Sprintf("\tWallets:\n")
	b += q.printWallets(q.walletRoot)

	b += fmt.Sprintf("\tSeed: %x\n\n", q.seed)

	b += fmt.Sprintf("\tParent: %x\n", q.parent)
	b += fmt.Sprintf("\tHeight: %x\n\n", q.height)
	return
}

// Encoded Variables:
//	non-nil siblings
//
//	germ
//	seed
//
//	cylinderTreeHead + entire tree (tbi)
func (q *Quorum) GobEncode() (gobQuorum []byte, err error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	// if q == nil, encode a zero quorum
	if q == nil {
		q = new(Quorum)
	}

	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	// Encode network variabes
	// Only encode non-nil siblings
	var encSiblings []*Sibling
	for _, s := range q.siblings {
		if s != nil {
			encSiblings = append(encSiblings, s)
		}
	}
	err = encoder.Encode(encSiblings)
	if err != nil {
		return
	}

	// Encode compile variables
	err = encoder.Encode(q.germ)
	if err != nil {
		return
	}
	err = encoder.Encode(q.seed)
	if err != nil {
		return
	}

	// Encode snap variables
	err = encoder.Encode(q.currentSnapshot)
	if err != nil {
		return
	}

	// Encode block tracking variables
	err = encoder.Encode(q.parent)
	if err != nil {
		return
	}
	err = encoder.Encode(q.height)
	if err != nil {
		return
	}

	gobQuorum = w.Bytes()
	println(len(gobQuorum))
	return
}

// Decoded Variables:
//	siblings
//
//	germ
//	seed
//
//  snap variables
func (q *Quorum) GobDecode(gobQuorum []byte) (err error) {
	// if q == nil, make a new quorum and decode into that
	if q == nil {
		q = new(Quorum)
	}

	r := bytes.NewBuffer(gobQuorum)
	decoder := gob.NewDecoder(r)

	// decode slice of siblings into the sibling array
	var encSiblings []*Sibling
	err = decoder.Decode(&encSiblings)
	if err != nil {
		return
	}
	for _, s := range encSiblings {
		q.siblings[s.index] = s
	}

	// decode compile variables
	err = decoder.Decode(&q.germ)
	if err != nil {
		return
	}
	err = decoder.Decode(&q.seed)
	if err != nil {
		return
	}

	// Decode snap variables
	err = decoder.Decode(&q.currentSnapshot)
	if err != nil {
		return
	}

	// Decode block tracking variables
	err = decoder.Decode(&q.parent)
	if err != nil {
		return
	}
	err = decoder.Decode(&q.height)
	if err != nil {
		return
	}

	return
}
