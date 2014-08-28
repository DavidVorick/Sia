package state

import (
	"errors"
	"io"

	"github.com/NebulousLabs/Sia/siacrypto"
)

// In setting these constants, remember that compensation weight can never
// exceed 2^32, since the number of atoms is measured using 32bit integers.
const (
	AtomSize         = 32   // In bytes
	AtomsPerSector   = 2048 // Eventually 2^16
	MaxUpdates       = 8    // Eventually 64
	MinConfirmations = 3    // Eventually 65
	MaxK             = 2    // Eventually 51
	MinK             = 1

	StandardK = 2
)

// Sector contains all the information about the sector of a wallet,
// including erasure information, potential future updates, and the total cost
// of the sector in its current state.
type Sector struct {
	// The number of atoms that have been allocated for the sector.
	Atoms uint16

	// The number of atoms that are currently allocated for uploading to the
	// sector.
	UpdateAtoms uint32

	// The minimum number of sibings in the quorum that need to remain
	// uncorrupted in order for the original data to be recoverable.
	K byte

	// The minimum number of siblings in the quorum that need to remain
	// uncorrupted in order for other pieces to be recoverable without using a
	// large amount of bandwidth.
	D byte

	// The hash of the hash set of the sector, where hash set is defined as an
	// ordered list of of hashes of each segment held by each sibling in the
	// quorum. Hash is kept as a variable so that there is a record in the
	// blockchain of what the exact appearance of the file should be.
	HashSet [QuorumSize]siacrypto.Hash

	// A list of all SectorModifiers active on each wallet. If the wallet
	// is not represented in the map, it only indicates that there are no
	// SectorModifiers active for that wallet. To check for a wallets
	// existence, one must transverse the wallet tree.  activeSectors
	// map[WalletID][]SectorModifier activeUploads map[UploadID]*Upload
	ActiveUpdates []SectorUpdate
}

// SectorFilename takes a wallet id and returns the filename of the sector
// associated with that wallet.
func (s *State) SectorFilename(id WalletID) (sectorFilename string) {
	sectorFilename = s.walletFilename(id) + ".sector"
	return
}

// SectorHash returns the combined hash of 'QuorumSize' Hashes.
func (s Sector) Hash() siacrypto.Hash {
	fullSet := make([]byte, siacrypto.HashSize*int(QuorumSize))
	for i := range s.HashSet {
		copy(fullSet[i*siacrypto.HashSize:], s.HashSet[i][:])
	}
	return siacrypto.HashBytes(fullSet)
}

// Helper function for merkle trees; takes two hashes, appends them, and then
// hashes their sum.
func joinHash(left, right siacrypto.Hash) siacrypto.Hash {
	return siacrypto.HashBytes(append(left[:], right[:]...))
}

// MerkleCollapse splits the provided data into segments of size AtomSize. It
// then recursively transforms these segments into a Merkle tree, and returns
// the root hash.
func MerkleCollapse(reader io.Reader, numAtoms uint16) (hash siacrypto.Hash, err error) {
	if numAtoms == 0 {
		err = errors.New("no data")
		return
	}
	if numAtoms == 1 {
		data := make([]byte, AtomSize)
		_, err = reader.Read(data)
		hash = siacrypto.HashBytes(data)
		return
	}

	// locate smallest power of 2 <= numAtoms
	var mid uint16 = 1
	for mid < numAtoms/2+numAtoms%2 {
		mid *= 2
	}

	// since we always read "left to right", no extra Seeking is necessary
	left, _ := MerkleCollapse(reader, mid)
	right, err := MerkleCollapse(reader, numAtoms-mid)
	hash = joinHash(left, right)
	return
}
