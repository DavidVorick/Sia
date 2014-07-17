// A List of Calls Available To Script:
// 1. Send
// 2. AddNewSibling
// 3. CreateWallet
package state

import (
	"siacrypto"
	"siaencoding"
	"siafiles"
)

const (
	QuorumSize        byte     = 4        // max siblings per quorum
	AtomSize          int      = 4096     // in bytes
	AtomsPerQuorum    int      = 16777216 // 64GB
	AtomsPerSector    uint16   = 200      // more causes DOS problems, is fixable. Final value likely to be 2^9-2^12
	BootstrapWalletID WalletID = 0
)

// The bootstrap script acts as a fountain, and can be called to spawn a new
// wallet with a minimal starting balance that will the economy to get flowing.
var BootstrapScript = []byte{
	0x27, 0x08, //       00 push 8 bytes of input (wallet id)
	0x01, 0x00, //       02 push 1 (high balance)
	0x02, 0xA8, 0x61, // 04 push 100 (low balance)
	0x2E, 0x01, //       08 read rest of input into buffer 1
	0x32, 0x01, //       10 call create wallet
	0xFF, //             12 exit
}

// the default script
// verifies public key, then transfers control to the input
func DefaultScript(publicKey siacrypto.PublicKey) []byte {
	return append([]byte{
		0x26, 0x10, 0x00, // 00 move data pointer to public key
		0x39, 0x20, 0x01, // 03 read public key into buffer 1
		0x2E, 0x02, //       06 read signed message into buffer 2
		0x34, 0x01, 0x02, // 08 verify signature
		0x38,             //             11 if invalid signature, reject
		0x26, 0x70, 0x00, // 12 move data pointer to input body
		0x2F, //             15 execute input
	}, publicKey[:]...)
}

type State struct {
	Metadata StateMetadata

	walletPrefix string
	wallets      uint32
	walletRoot   *walletNode

	eventRoot *eventNode
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
	filename = s.walletPrefix + suffix
	return
}

// NewBootstrapState initializes the quorum so that there is a bootstrap wallet
// which has funds. This wallet can then be used by other potential siblings as
// a fountain, and the network has some way to get the ball rolling.
func (s *State) BootstrapState(sib *Sibling) (err error) {
	// Create the bootstrap wallet, which acts as a fountain to get the economy
	// started.
	w := Wallet{
		ID:      BootstrapWalletID,
		Balance: NewBalance(0, 25000000),
		Script:  BootstrapScript,
	}
	err = s.InsertWallet(w)
	if err != nil {
		return
	}

	// Create a walle with the default script for the sibling to use.
	defaultScript := DefaultScript(sib.PublicKey)
	sibWallet := &Wallet{
		ID:      sib.WalletID,
		Balance: NewBalance(0, 1000000),
		Script:  defaultScript,
	}
	s.AddSibling(sibWallet, sib)

	return
}
