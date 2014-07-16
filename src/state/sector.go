package state

import (
	//"io"
	//"os"
	"siacrypto"
)

type SectorSettings struct {
	Atoms uint16
	K     byte
	Hash  siacrypto.Hash
}

/*
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
	walletFilename := s.walletFilename(id)
	sectorFilename = walletFilename + ".sector"
	return
}

func (s *State) BuildStorageProof(id WalletID, atomIndex int) (proofStack []*siacrypto.Hash) {
	return
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

func (s *State) VerifyStorageProof(id WalletID, atomIndex uint16, sibling byte, proofBase []byte, proofStack []*siacrypto.Hash) bool {
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

	//w := q.LoadWallet(id)

	// determine that proofStack is long enough
	//var proofsNeeded int
	//var proofCounter uint32
	//for proofCounter<<1 <= uint32(w.sectorAtoms) {
	//	proofsNeeded += 1
	//	proofCounter <<= 1
	//}
	//if len(proofStack) < proofsNeeded {
	//	return false
	//}

	// build the hash up from the base
	//initialHash := siacrypto.CalculateHash(proofBase)
	//initialHigh := w.sectorAtoms
	//finalHash := buildMerkleRoot(initialHigh, atomIndex, initialHash, proofStack)

	//if finalHash != expectedHash {
	//return false
	//}
	return true
}*/
