package state

const (
	// QuorumSize is the maximum number of siblings in a quorum.
	QuorumSize byte = 4

	// AtomsPerQuorum is the maximum number of atoms that can be stored on
	// a single quorum.
	AtomsPerQuorum int = 16777216
)

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
}

// AtomsInUse returns the number of atoms being consumed by the whole quorum.
func (s *State) AtomsInUse() int {
	return s.walletRoot.weight
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

// Initialize takes all the siblings and sets them to inactive
func (s *State) Initialize() {
	for i := range s.Metadata.Siblings {
		s.Metadata.Siblings[i].Status = ^byte(0)
	}
}
