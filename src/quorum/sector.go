package quorum

import (
	"io"
	"siacrypto"
)

// SectorFilename takes a wallet id and returns the filename of the sector
// associated with that wallet.
func (q *Quorum) SectorFilename(id WalletID) (sectorFilename string) {
	walletFilename := q.walletFilename(id)
	sectorFilename = walletFilename + ".sector"
	return
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
func (q *Quorum) MerkleCollapse(reader io.Reader) (hash siacrypto.Hash) {
	// Loop through every atom in the reader, building out the Merkle Tree in
	// linear time.
	prevHashes := make([]*siacrypto.Hash, 1) // prevHashes starts at size 1, to store the first hash
	atom := make([]byte, AtomSize)
	var atoms int
	for _, err := reader.Read(atom); err == nil; atoms++ {
		// If atoms is a power of 2, increase the length of prevHashes
		if atoms&(atoms-1) == 0 {
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
		copy(prevHashes[i][:], hash[:])
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
