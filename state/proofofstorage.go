package state

import (
	"errors"
	"io"
	"os"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
)

type StorageProof struct {
	AtomBase  [AtomSize]byte
	HashStack []*siacrypto.Hash
}

var ErrEmptyQuorum = errors.New("quorum is not storing any data")

// Proof location uses s.Metadata.PoStorageSeed to determine which atom of
// which wallet is being checked for during proof of storage.
func (s *State) proofLocation() (id WalletID, index uint16, err error) {
	// Take the first 8 bytes of the storage proof and convert to a uint64 - a
	// random number.
	seedInt := siaencoding.DecUint64(s.Metadata.PoStorageSeed[:])

	// Can't take the modulus of 0
	if s.walletRoot.weight == 0 {
		err = ErrEmptyQuorum
		return
	} else {
		seedInt %= uint64(s.walletRoot.weight)
	}

	// Get the node and index associated with the seed weight.
	node, index, err := s.weightNode(seedInt)
	if err != nil {
		return
	}
	id = node.id
	return
}

// buildProof constructs a list of hashes using the following procedure. The
// storage proof requires traversing the Merkle tree from the proofIndex node
// to the root. On each level of the tree, we must provide the hash of "sister"
// node. (Since this is a binary tree, the sister node is the other node with
// the same parent as us.) To obtain this hash, we call MerkleCollapse on the
// segment of data corresponding to the sister. This segment will double in
// size on each iteration until we reach the root.
func buildProof(rs io.ReadSeeker, numAtoms, proofIndex uint16) (sp StorageProof, err error) {
	// get AtomBase
	if _, err = rs.Seek(int64(proofIndex)*int64(AtomSize), 0); err != nil {
		return
	}
	if _, err = rs.Read(sp.AtomBase[:]); err != nil {
		return
	}

	// sisterIndex helper function:
	//   if the sector is divided into segments of length 'size' and
	//   grouped pairwise, then proofIndex lies inside a segment
	//   that is one half of a pair. sisterIndex returns the index
	//   where the other half begins.
	//   e.g.: (5, 1) -> 4, (5, 2) -> 6, (5, 4) -> 0, ...
	sisterIndex := func(index, size uint16) uint16 {
		if index%(size*2) < size { // left child or right child?
			return (index/size + 1) * size
		} else {
			return (index/size - 1) * size
		}
	}

	// calculate hashes of each sister
	for size := uint16(1); size < numAtoms; size <<= 1 {
		// determine index
		i := sisterIndex(proofIndex, size)
		if i >= numAtoms {
			// append dummy hash
			sp.HashStack = append(sp.HashStack, nil)
			continue
		}

		// seek to beginning of segment
		rs.Seek(int64(i)*int64(AtomSize), 0)

		// truncate number of atoms to read, if necessary
		truncSize := size
		if i+size > numAtoms {
			truncSize = numAtoms - i
		}

		// calculate and append hash
		var hash siacrypto.Hash
		hash, err = MerkleCollapse(rs, truncSize)
		if err != nil {
			return
		}
		sp.HashStack = append(sp.HashStack, &hash)
	}

	return
}

// BuildStorageProof is a simple wrapper around buildProof.
func (s *State) BuildStorageProof() (sp StorageProof, err error) {
	// Get the wallet and atom being proven for.
	walletID, proofIndex, err := s.proofLocation()
	if err != nil {
		return
	}

	// read the sector data
	sectorFilename := s.SectorFilename(walletID)
	file, err := os.Open(sectorFilename)
	if err != nil {
		return
	}
	defer file.Close()

	// determine numAtoms
	w, err := s.LoadWallet(walletID)
	if err != nil {
		return
	}
	numAtoms := w.Sector.Atoms

	sp, err = buildProof(file, numAtoms, proofIndex)
	if err != nil {
		return
	}
	return
}

// foldHashes traverses a proofStack, hashing elements together to produce the
// root-level hash. Care must be taken to ensure that the correct ordering is
// used when concatenating hashes.
func foldHashes(sp StorageProof, proofIndex uint16) (h siacrypto.Hash) {
	h = siacrypto.HashBytes(sp.AtomBase[:])

	var size uint16 = 1
	for i := 0; i < len(sp.HashStack); i, size = i+1, size*2 {
		// skip dummy hashes
		if sp.HashStack[i] == nil {
			continue
		}
		if proofIndex%(size*2) < size { // base is on the left branch
			h = joinHash(h, *sp.HashStack[i])
		} else {
			h = joinHash(*sp.HashStack[i], h)
		}
	}

	return
}

// VerifyStorageProof verifies that a specified atom, along with a
// corresponding proofStack, can be used to reconstruct the original root
// Merkle hash.
func (s *State) VerifyStorageProof(sibling byte, sp StorageProof) (verified bool, err error) {
	// Get the wallet & atom index of the bytes being proven for.
	walletID, proofIndex, err := s.proofLocation()
	if err != nil {
		return
	}

	// Get the expected hash from the wallet.
	w, err := s.LoadWallet(walletID)
	if err != nil {
		return
	}
	expectedHash := w.Sector.HashSet[sibling]

	// build the hash up from the base
	verified = foldHashes(sp, proofIndex) == expectedHash
	return
}
