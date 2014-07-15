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

// This is the prefix that the state will use when opening wallets as files.
// Eventually, logic will be implemented to move all of the wallets and files
// if the prefex is changed.
func (s *State) SetWalletPrefix(walletPrefix string) {
	s.walletPrefix = walletPrefix
}

func (s *State) walletFilename(id WalletID) (filename string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	filename = q.walletFilenamePrefix + suffix
	return
}
