package quorum

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
	// quorum-wide lock
	lock sync.RWMutex

	// Network Variables
	siblings [QuorumSize]*Sibling

	// Compile Variables
	germ Entropy // Where internal entropy is stored before external entropy is applied
	seed Entropy // Used to generate random numbers during compilation

	// wallet management
	walletPrefix string
	numNodes     uint32
	walletRoot   *walletNode

	// snapshot management
	snapSize              [2]uint32
	snapWalletSliceOffset [2]uint32
	snapWallets           [2]uint32
	currentSnap           bool // false == snap0, true == snap1
}

// Getter for the siblings private variable
func (q *Quorum) Siblings() [QuorumSize]*Sibling {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.siblings
}

// q.Status() enumerates the variables of the quorum in a human-readable output
func (q *Quorum) Status() (b string) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	b = "\nQuorum Status:\n"

	b += fmt.Sprintf("\tSiblings:\n")
	for _, s := range q.siblings {
		if s != nil {
			pubKeyHash, err := s.publicKey.Hash()
			if err != nil {
				// ???
			}

			b += fmt.Sprintf("\t\t%v\n", s.index)
			b += fmt.Sprintf("\t\t\tAddress: %v\n", s.address)
			b += fmt.Sprintf("\t\t\tPublic Key: %v\n", pubKeyHash[:6])
		}
	}
	b += fmt.Sprintf("\n")

	b += fmt.Sprintf("\tWallets:\n")
	b += q.printWallets(q.walletRoot)

	b += fmt.Sprintf("\tSeed: %x\n\n", q.seed)
	return
}

// This is the prefix that the quorum will use when opening wallets as files.
// There is no getter, because one is not seen as necessary.
func (q *Quorum) SetWalletPrefix(walletPrefix string) {
	q.walletPrefix = walletPrefix
}

func (q *Quorum) GetWalletPrefix() string {
	return q.walletPrefix
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
	err = encoder.Encode(q.snapSize[0])
	if err != nil {
		return
	}
	err = encoder.Encode(q.snapSize[1])
	if err != nil {
		return
	}
	err = encoder.Encode(q.snapWalletSliceOffset[0])
	if err != nil {
		return
	}
	err = encoder.Encode(q.snapWalletSliceOffset[1])
	if err != nil {
		return
	}
	err = encoder.Encode(q.snapWallets[0])
	if err != nil {
		return
	}
	err = encoder.Encode(q.snapWallets[1])
	if err != nil {
		return
	}
	err = encoder.Encode(q.currentSnap)
	if err != nil {
		return
	}

	gobQuorum = w.Bytes()
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
	err = decoder.Decode(&q.snapSize[0])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.snapSize[1])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.snapWalletSliceOffset[0])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.snapWalletSliceOffset[1])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.snapWallets[0])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.snapWallets[1])
	if err != nil {
		return
	}
	err = decoder.Decode(&q.currentSnap)
	if err != nil {
		return
	}

	return
}
