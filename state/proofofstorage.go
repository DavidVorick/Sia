package state

import (
	"io"
	"os"

	"github.com/NebulousLabs/Sia/siacrypto"
)

type StorageProof struct {
	Base       [AtomSize]byte
	ProofStack []siacrypto.Hash
}

// buildProof constructs a list of hashes using the following procedure. The
// storage proof requires traversing the Merkle tree from the proofIndex node
// to the root. On each level of the tree, we must provide the hash of "sister"
// node. (Since this is a binary tree, the sister node is the other node with
// the same parent as us.) To obtain this hash, we call MerkleCollapse on the
// segment of data corresponding to the sister. This segment will double in
// size on each iteration until we reach the root.
func buildProof(rs io.ReadSeeker, numAtoms, proofIndex uint16) (proofBytes []byte, proofStack []*siacrypto.Hash) {
	// get proofBytes
	_, err := rs.Seek(int64(proofIndex)*int64(AtomSize), 0)
	if err != nil {
		panic(err)
	}
	proofBytes = make([]byte, AtomSize)
	_, err = rs.Read(proofBytes)
	if err != nil {
		panic(err)
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
			proofStack = append(proofStack, nil)
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
		hash, err := MerkleCollapse(rs, truncSize)
		if err != nil {
			panic(err)
		}
		proofStack = append(proofStack, &hash)
	}

	return
}

// BuildStorageProof is a simple wrapper around buildProof.
func (s *State) BuildStorageProof(id WalletID, proofIndex uint16) (proofBytes []byte, proofStack []*siacrypto.Hash) {
	// read the sector data
	sectorFilename := s.SectorFilename(id)
	file, err := os.Open(sectorFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// determine numAtoms
	w, err := s.LoadWallet(id)
	if err != nil {
		panic(err)
	}
	numAtoms := w.SectorSettings.Atoms

	return buildProof(file, numAtoms, proofIndex)
}

// foldHashes traverses a proofStack, hashing elements together to produce the
// root-level hash. Care must be taken to ensure that the correct ordering is
// used when concatenating hashes.
func foldHashes(base siacrypto.Hash, proofIndex uint16, proofStack []*siacrypto.Hash) (h siacrypto.Hash) {
	h = base

	var size uint16 = 1
	for i := 0; i < len(proofStack); i, size = i+1, size*2 {
		// skip dummy hashes
		if proofStack[i] == nil {
			continue
		}
		if proofIndex%(size*2) < size { // base is on the left branch
			h = joinHash(h, *proofStack[i])
		} else {
			h = joinHash(*proofStack[i], h)
		}
	}

	return
}

// VerifyStorageProof verifies that a specified atom, along with a
// corresponding proofStack, can be used to reconstruct the original root
// Merkle hash.
// TODO: think about removing this function or combining it with foldHashes
func (s *State) VerifyStorageProof(id WalletID, proofIndex uint16, sibling byte, proofBase []byte, proofStack []*siacrypto.Hash) bool {
	// get the intended hash from the segment stored on disk
	sectorFilename := s.SectorFilename(id)
	file, err := os.Open(sectorFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// seek to the location where this particular siblings hash is stored
	hashLocation := int64(sibling) * int64(siacrypto.HashSize)
	var expectedHash siacrypto.Hash
	_, err = file.ReadAt(expectedHash[:], hashLocation)
	if err != nil {
		panic(err)
	}

	// build the hash up from the base
	initialHash := siacrypto.HashBytes(proofBase)
	finalHash := foldHashes(initialHash, proofIndex, proofStack)

	return finalHash == expectedHash
}
