package state

import (
	"io"
	"os"
	"siacrypto"
)

type SectorSettings struct {
	// The number of atoms that have been allocated for the sector.
	Atoms uint16

	// The number of atoms that are currently allocated for uploading to the
	// sector.
	UploadAtoms uint16

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
	Hash siacrypto.Hash
}

type SectorModifier interface {
	Hash() siacrypto.Hash
}

func (s *State) ActiveParentHash(w Wallet, parentHash siacrypto.Hash) bool {
	modifiers, exists := s.activeUploads[w.ID]
	if exists {
		latestModifier := modifiers[len(modifiers)-1]
		return parentHash == latestModifier.Hash()
	} else {
		return parentHash == w.SectorSettings.Hash
	}
}

func (s *State) AppendSectorModifier(id WalletID, sm SectorModifier) {
	s.activeUploads[id] = append(s.activeUploads[id], sm)
}

// MerkleCollapse takes a reader as input and treats each set of AtomSize bytes
// as an atom. It then creates a Merkle Tree of the atoms. The algorithm for
// doing this keeps in memory the previous hash at each level. Each atom is
// hashed as level one, but if there is already a hash stored at level one,
// then the hash is taken again producing a level two hash. If there is already
// a hash at level two, the hash is taken again... and so on. Once all the
// atoms have been read, a cleanup loop is used to append the highest level
// hash to each lower level hash and arrive at the final value. See the
// whitepaper for a specification of how to construct Merkle Trees from a
// sector. This algorithm takes linear time and logarithmic space.
func MerkleCollapse(reader io.Reader) (hash siacrypto.Hash) {
	// Loop through every atom in the reader, building out the Merkle Tree in
	// linear time.
	prevHashes := make([]*siacrypto.Hash, 0)
	atom := make([]byte, AtomSize)
	var atoms int
	for _, err := reader.Read(atom); err == nil; atoms++ {
		// If atoms is a power of 2, increase the length of prevHashes
		if (atoms+1)&(atoms) == 0 {
			prevHashes = append(prevHashes, nil)
		}

		// Take the hash of the current atom and merge it with the hashes in
		// prevHashes, one level at a time until an empty slot in prevHashes is
		// found. Each time a merge happens at a level, the level is reset to nil.
		hash := siacrypto.CalculateHash(atom)
		var i int
		for i = 0; prevHashes[i] != nil; i++ {
			hash = siacrypto.CalculateHash(append(prevHashes[i][:], hash[:]...))
			prevHashes[i] = nil
		}

		// store the new hash in the first empty slot
		prevHashes[i] = new(siacrypto.Hash)
		copy(prevHashes[i][:], hash[:])
		_, err = reader.Read(atom)
	}

	// check that at least something was read
	if len(prevHashes) == 0 {
		return
	}

	// Merge the hashes into the final tree, starting with the highest level hash
	// and working down to the lowest.
	hash = *prevHashes[len(prevHashes)-1]
	for i := len(prevHashes) - 2; i >= 0; i-- {
		if prevHashes[i] != nil {
			hash = siacrypto.CalculateHash(append(hash[:], prevHashes[i][:]...))
		}
	}
	return
}

func SectorHash(hashSet [QuorumSize]siacrypto.Hash) siacrypto.Hash {
	atomRepresentation := make([]byte, AtomSize) // regardless of quorumsize, must hash a whole atom
	for i := range hashSet {
		copy(atomRepresentation[i*siacrypto.HashSize:], hashSet[i][:])
	}
	return siacrypto.CalculateHash(atomRepresentation)
}

// SectorFilename takes a wallet id and returns the filename of the sector
// associated with that wallet.
func (s *State) SectorFilename(id WalletID) (sectorFilename string) {
	sectorFilename = s.walletFilename(id) + ".sector"
	return
}

// BuildStorageProof constructs a list of hashes using the following procedure.
// The storage proof requires traversing the Merkle tree from the proofIndex node to the root.
// On each level of the tree, we must provide the hash of "sister" node.
// (Since this is a binary tree, the sister node is the other node with the same parent as us.)
// To obtain this hash, we call MerkleCollapse on the segment of data corresponding to the sister.
// This segment will double in size on each iteration until we reach the root.
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
			continue
		}

		// create a bounded reader via Seek and LimitReader
		// this negates the need for any special processing of imperfectly balanced trees
		_, err = rs.Seek(int64(i)*int64(AtomSize), 0)
		if err != nil {
			panic(err)
		}
		r := io.LimitReader(rs, int64(size))

		// calculate and append hash
		hash := MerkleCollapse(r)
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

// Recursive algorithm to take a list of proofs that fall down a merkle tree,
// get to the bottom, and then hash them in the right order to build back up to
// the merkle root
func buildMerkleRoot(high uint16, search uint16, base siacrypto.Hash, proofStack []*siacrypto.Hash) siacrypto.Hash {
	// base case: does high equal 1 or 0?
	if high == 1 {
		if search == 0 {
			return siacrypto.CalculateHash(append(base[:], proofStack[0][:]...))
		} else {
			return siacrypto.CalculateHash(append(proofStack[0][:], base[:]...))
		}
	}
	if high == 0 {
		return base
	}

	// find the highest power of 2 that fits into 'high' (but not completely, so 2^2 for 8, 2^3 for 9)
	var divider uint16
	for divider<<1 < high {
		divider <<= 1
	}

	if search < divider {
		nextHash := buildMerkleRoot(divider, search-(high-divider), base, proofStack[1:])
		return siacrypto.CalculateHash(append(nextHash[:], proofStack[0][:]...))
	} else {
		nextHash := buildMerkleRoot(high-divider, search, base, proofStack[1:])
		return siacrypto.CalculateHash(append(proofStack[0][:], nextHash[:]...))
	}
}

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
	_, err = file.Seek(hashLocation, 0)
	if err != nil {
		panic(err)
	}
	var expectedHash siacrypto.Hash
	_, err = file.Read(expectedHash[:])
	if err != nil {
		panic(err)
	}

	w, err := s.LoadWallet(id)
	if err != nil {
		panic(err)
	}

	// determine that proofStack is long enough
	var proofsNeeded int
	var proofCounter uint32
	for proofCounter<<1 <= uint32(w.SectorSettings.Atoms) {
		proofsNeeded += 1
		proofCounter <<= 1
	}
	if len(proofStack) < proofsNeeded {
		return false
	}

	// build the hash up from the base
	initialHash := siacrypto.CalculateHash(proofBase)
	initialHigh := w.SectorSettings.Atoms
	finalHash := buildMerkleRoot(initialHigh, proofIndex, initialHash, proofStack)

	if finalHash != expectedHash {
		return false
	}
	return true
}
