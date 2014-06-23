package quorum

import (
	"os"
	"siacrypto"
)

func (q *Quorum) SectorFilename(id WalletID) (sectorFilename string) {
	walletFilename := q.walletFilename(id)
	sectorFilename = walletFilename + ".sector"
	return
}

func (q *Quorum) MerkleCollapse(filename string) (hash siacrypto.Hash, err error) {
	// open the file, get the size of the file
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	info, err := file.Stat()
	if err != nil {
		return
	}
	size := info.Size()

	// find the number of layers needed to build the merkle tree
	size /= int64(AtomSize)
	var numAtoms int
	for size != 0 {
		size >>= 1
		numAtoms++
	}

	prevHashes := make([]*siacrypto.Hash, numAtoms)
	atom := make([]byte, 4096)
	for size = info.Size(); size > 0; size -= 4096 {
		_, err = file.Read(atom)
		if err != nil {
			return
		}

		hash := siacrypto.CalculateHash(atom)
		var i int
		for i = 0; prevHashes[i] != nil; i++ {
			hash = siacrypto.CalculateHash(append(prevHashes[i][:], hash[:]...))
			prevHashes[i] = nil
		}
		copy(prevHashes[i][:], hash[:])
	}

	hash = *prevHashes[numAtoms-1]
	for i := numAtoms - 2; i >= 0; i-- {
		if prevHashes[i] != nil {
			hash = siacrypto.CalculateHash(append(hash[:], prevHashes[i][:]...))
		}
	}
	return
}
