package state

import (
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/siafiles"
)

const (
	// QuorumSize is the maximum number of siblings in a quorum.
	QuorumSize byte = 4
	// AtomSize is the number of bytes in an atom.
	AtomSize int = 32
	// AtomsPerQuorum is the maximum number of atoms that can be stored on
	// a single quorum.
	AtomsPerQuorum int = 16777216
	// AtomsPerSector is the number of atoms in a sector. Final value
	// likely to be 2^13-2^16
	AtomsPerSector uint16 = 1024

	// SiblingPassiveWindow is the number of blocks that a sibling is
	// allowed to be passive.
	SiblingPassiveWindow = 5
)

// A Sibling is the public facing information of participants on the quorum.
// Every quorum contains a list of all siblings. The Status of a sibling
// indicates it's standing with the quorum. ^byte(0) indicates that the sibling
// is 'Inactive', and that there are no hosts filling that position. A standing
// of '5-1' indicates that the sibling is 'Passive', with the number indicating
// how many compiles until the sibling becomes active. A passive sibling is
// sent updates during consensus, and its signatures are accepted during
// consensus, but its updates are not required. Updates are ignored and a
// passive sibling will not be included in compensation. An active sibling is a
// full sibing that _must_ participate in consensus and provide updates to the
// network.
type Sibling struct {
	Status    byte
	Index     byte
	Address   network.Address
	PublicKey siacrypto.PublicKey
	WalletID  WalletID
}

// Active returns true if the sibling is a fully active member of the quorum
// according to the status variable, false if the sibling is passive or
// inactive.
func (sib Sibling) Active() bool {
	return sib.Status == 0
}

// Inactive returns true if the sibling is inactive, and retuns false if the
// sibling is active or passive.
func (sib Sibling) Inactive() bool {
	return sib.Status == ^byte(0)
}

// The State struct contains all of the information about the current state of
// the quorum. This includes the list of wallets, all events, any file-updates
// that are in progress, and eventually information about the metaquorums as
// well.
type State struct {
	// A struct containing all of the simple, single-variable data of the quorum.
	Metadata Metadata

	// All of the wallet data on the quorum, including information on how to
	// read the wallet segments from disk. 'wallets' indicats the number of
	// wallets in the State, and is placed for convenience. This number could
	// also be derived by doing a search starting at the walletRoot.
	walletPrefix string
	wallets      uint32
	walletRoot   *walletNode

	// Points to the skip list that contains all of the events.
	eventRoot *eventNode

	// Maintains a list of all SectorModifiers active on each wallet. If the
	// wallet is not represented in the map, it only indicates that there are
	// no SectorModifiers active for that wallet. To check for a wallets
	// existence, one must transverse the wallet tree.
	// activeSectors map[WalletID][]SectorModifier
	// activeUploads map[UploadID]*Upload
	activeUpdates map[WalletID][]SectorUpdate
}

// SetWalletPrefix is a setter that sets the state walletPrefix field.
// TODO: move this documentation to package docstring?
// This is the prefix that the state will use when opening wallets as files.
// Eventually, logic will be implemented to move all of the wallets and files
// if the prefex is changed. It is permissible to change the wallet prefix in
// the middle of operation.
func (s *State) SetWalletPrefix(walletPrefix string) {
	// Though the header says it's admissible, that isn't actually supported in
	// the current implementation. But it's on the todo list.

	s.walletPrefix = walletPrefix
}

// walletFilename returns the filename for a wallet, receiving only the id of
// the wallet as input.
func (s *State) walletFilename(id WalletID) (filename string) {
	// Turn the id into a suffix that will follow the quorum prefix
	suffixBytes := siaencoding.EncUint64(uint64(id))
	suffix := siafiles.SafeFilename(suffixBytes)
	filename = s.walletPrefix + "." + suffix
	return
}

// AtomsInUse returns the number of atoms being consumed by the whole quorum.
func (s *State) AtomsInUse() int {
	return s.walletRoot.weight
}

// TossSibling removes a sibling from the list of siblings.
func (s *State) TossSibling(i byte) {
	s.Metadata.Siblings[i] = *new(Sibling)
}
