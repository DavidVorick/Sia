package state

import (
	"errors"
	"io"
	"os"

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

// SectorSettings contains all the information about the sector of a wallet,
// including erasure information, potential future updates, and the total cost
// of the sector in its current state.
type SectorSettings struct {
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
func (s SectorSettings) Hash() siacrypto.Hash {
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
