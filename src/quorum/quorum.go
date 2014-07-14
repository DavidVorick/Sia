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
	QuorumSize     byte   = 4        // max siblings per quorum
	AtomSize       int    = 4096     // in bytes
	AtomsPerQuorum int    = 16777216 // 64GB
	AtomsPerSector uint16 = 200      // more causes DOS problems, is fixable. Final value likely to be 2^9-2^12
)

type State struct {
	metaData QuorumMetadata

	walletPrefix string
	wallets      uint32
	walletRoot   *walletNode

	eventRoot    *eventNode
}

func (q *Quorum) Init() {
	q.uploads = make(map[WalletID][]*upload)
	q.storagePrice = NewBalance(0, 1)
}

// This is the prefix that the quorum will use when opening wallets as files.
// Eventually, logic will be implemented to move all of the wallets and files
// if the prefex is changed.
func (q *Quorum) SetWalletPrefix(walletPrefix string) {
	q.walletPrefix = walletPrefix
}

func (s *State) walletFilename(id WalletID) (filename string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	filename = q.walletFilenamePrefix + suffix
	return
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
			b += fmt.Sprintf("\t\t\tPublic Key: %x\n", pubKeyHash)
		}
	}
	b += fmt.Sprintf("\n")

	b += fmt.Sprintf("\tWallets:\n")
	b += q.printWallets(q.walletRoot)

	b += fmt.Sprintf("\tSeed: %x\n\n", q.seed)

	if q.walletRoot != nil {
		b += fmt.Sprintf("\tWeight: %x\n", q.walletRoot.weight)
	} else {
		b += fmt.Sprintf("\tWeight: 0\n")
	}
	// b += fmt.Sprintf("\tParent: %x\n", q.parent)
	b += fmt.Sprintf("\tHeight: %x\n\n", q.height)
	return
}
