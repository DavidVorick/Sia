// A List of Calls Available To Script:
// 1. Send
// 2. AddNewSibling
// 3. CreateWallet
package state

import (
	"network"
	"siacrypto"
	"siaencoding"
	"siafiles"
)

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings.
type Sibling struct {
	Active    bool
	Index     byte
	Address   network.Address
	PublicKey siacrypto.PublicKey
	WalletID  WalletID
}

const (
	QuorumSize     byte   = 4        // max siblings per quorum
	AtomSize       int    = 64       // in bytes
	AtomsPerQuorum int    = 16777216 // 1GB
	AtomsPerSector uint16 = 1024     // more causes DOS problems, is fixable. Final value likely to be 2^13-2^16
)

type State struct {
	// A struct containing all of the simple, single-variable data of the quorum.
	Metadata StateMetadata

	// All of the wallet data on the quorum, including information on how to read
	// the wallet segments from disk. 'wallets' indicats the number of wallets in
	// the State, and is placed for convenience. This number could also be
	// derived by doing a search starting at the walletRoot.
	walletPrefix string
	wallets      uint32
	walletRoot   *walletNode

	// Points to the skip list that contains all of the events.
	eventRoot *eventNode

	// Maintains a list of all SectorModifiers active on each wallet. If the
	// wallet is not represented in the map, it only indicates that there are no
	// SectorModifiers active for that wallet. To check for a wallets existence,
	// one must transverse the wallet tree.
	activeSectors map[WalletID][]SectorModifier
	activeUploads map[UploadID]*Upload
}

// This is the prefix that the state will use when opening wallets as files.
// Eventually, logic will be implemented to move all of the wallets and files
// if the prefex is changed. It is permissible to change the wallet prefix in
// the middle of operation.
func (s *State) SetWalletPrefix(walletPrefix string) {
	// Though the header says it's admissible, that isn't actually supported in
	// the current implementation. But it's on the todo list.

	s.walletPrefix = walletPrefix
}

func (s *State) walletFilename(id WalletID) (filename string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	filename = s.walletPrefix + suffix
	return
}

func (s *State) Weight() int {
	return s.walletRoot.weight
}

// Removes a sibling from the list of siblings
func (s *State) TossSibling(i byte) {
	s.Metadata.Siblings[i] = *new(Sibling)
}

func (s *State) ActiveUpload(uid UploadID) (upload Upload, exists bool) {
	uploadPointer, exists := s.activeUploads[uid]
	upload = *uploadPointer
	return
}
