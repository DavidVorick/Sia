package delta

import (
	"errors"
	"state"
)

const (
	CreateWalletCost = 8
	SendCost         = 6
	AddSiblingCost   = 500
)

var (
	cwerrInsufficientBalance = errors.New("Insufficient balance to create a wallet with the given balance.")

	aserrNoEmptySiblings = errors.New("There are no empty spots in the quorum.")
)

// CreateWallet takes an id, a Balance, and an initial script and uses
// those to create a new wallet that gets stored in stable memory.
// If a wallet of that id already exists then the process aborts.
func (e *Engine) CreateWallet(w *state.Wallet, childID state.WalletID, childBalance state.Balance, childScript []byte) (err error) {
	// Check that the wallet making the call has enough funds to deposit into the
	// wallet being created, and then subtract the funds from the parent wallet.
	if w.Balance.Compare(childBalance) < 0 {
		err = cwerrInsufficientBalance
		return
	}
	w.Balance.Subtract(childBalance)

	// Create a new wallet based on the inputs.
	childWallet := state.Wallet{
		ID:      childID,
		Balance: childBalance,
		Script:  childScript,
	}

	// Save the child wallet.
	err = e.state.SaveWallet(childWallet)
	return
}

// Currently, AddSibling tries to add the new sibling to the existing quorum
// and throws the sibling out if there's no space. Once quorums are
// communicating, the AddSibling routine will always succeed.
func (e *Engine) AddSibling(w *state.Wallet, sib state.Sibling) (err error) {
	// first check that the wallet can afford the down payment.

	// Look through the quorum for an empty sibling.
	for i := byte(0); i < state.QuorumSize; i++ {
		if !e.state.Metadata.Siblings[i].Active {
			sib.Active = true
			sib.Index = i
			sib.WalletID = w.ID
			e.state.Metadata.Siblings[i] = sib
			break
		}
	}

	if !sib.Active {
		err = aserrNoEmptySiblings
		return
	}

	// Charge the wallet some volume that's required as a down payment.

	return
}

/*
func (s *State) Send(w *Wallet, amount Balance, destID WalletID) (cost int, err error) {
	cost += 1
	if w.Balance.Compare(amount) < 0 {
		err = errors.New("insufficient balance")
		return
	}
	cost += 2
	destWallet := s.LoadWallet(destID)
	if destWallet == nil {
		err = errors.New("destination wallet does not exist")
		return
	}

	cost += 3
	w.Balance.Subtract(amount)
	destWallet.Balance.Add(amount)
	s.SaveWallet(destWallet)
	return
}

// Every wallet has a single sector, which can be up to 2^16 atoms of 4kb each,
// or 32GB total with 0 redundancy. Wallets pay for the size of their sector.
func (s *State) ResizeSectorErase(w *Wallet, atoms uint16, k byte) (cost int, weight int, err error) {
	cost += 3
	weightDelta := int(atoms)
	// weightDelta -= int(w.sectorAtoms)
	if weightDelta == 0 {
		return
	}

	// update the weights in the wallet tree
	s.updateWeight(w.ID, weightDelta)
	if s.walletRoot.weight > AtomsPerQuorum {
		s.updateWeight(w.ID, -weightDelta)
		return
	}
	weight = weightDelta

	// remove the file and return if the sector has been resized to length 0
	walletName := s.walletFilename(w.ID)
	sectorName := walletName + ".sector"
	if atoms == 0 {
		os.Remove(sectorName)
		return
	}

	// derive the name of the file housing the sector, and truncate the file
	file, err := os.Create(sectorName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// extend the file to being to proper length
	err = file.Truncate(int64(atoms) * int64(AtomSize))
	if err != nil {
		panic(err)
	}

	// update the hash associated with the sector
	_, err = file.Seek(int64(AtomSize), 0) // first atom contains hash information
	if err != nil {
		panic(err)
	}
	zeroMerkle := MerkleCollapse(file)

	// build the first atom of the file to contain all of the hashes
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	for i := byte(0); i < QuorumSize; i++ {
		_, err := file.Write(zeroMerkle[:])
		if err != nil {
			panic(err)
		}
	}

	// get the hash of the first atom as the sector hash
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	firstAtom := make([]byte, AtomSize)
	_, err = file.Read(firstAtom)
	if err != nil {
		panic(err)
	}
	// w.sectorAtoms = atoms
	// w.sectorM = k
	// w.sectorHash = siacrypto.CalculateHash(firstAtom)

	return
}

type UploadArgs struct {
	ParentHash    siacrypto.Hash
	NewHashSet    [QuorumSize]siacrypto.Hash
	AtomsChanged  uint16
	Confirmations byte
	Deadline      uint32
}

// First sectors are allocated, and then changes are uploaded to them. This
// creates a change.
func (s *State) ProposeUpload(w *Wallet, parentHash siacrypto.Hash, newHashSet [QuorumSize]siacrypto.Hash, atomsChanged uint16, confirmations byte, deadline uint32) (cost int, weight uint16, err error) {
	cost += 2

	// make sure the sector is allocated
	//if w.sectorAtoms == 0 {
	//		err = errors.New("Sector is not allocated")
	//		return
	//	}

	// make sure that the confirmations value is a reasonable value
	if confirmations > QuorumSize {
		err = errors.New("confirmations cannot be greater than quorum size")
		return
	}
	//if confirmations < w.sectorM {
	//	err = errors.New("confirmations cannot be less than the value of 'm' for the given sector")
	//		return
	//	}

	// make sure the deadline is a reasonable value
	if deadline > MaxDeadline+q.height {
		err = errors.New("deadline is too far in the future")
		return
	}
	if deadline <= q.height {
		err = errors.New("deadline has already arrived")
		return
	}

	cost += 2
	// look up all of the open uploads on this sector, and compare their hashes
	// to the parent hash of this upload. As soon as one is found (potentially
	// starting directly from the existing hash), all remaining uploads are
	// truncated. There can only exist a single chain of potential uploads, all
	// other get defeated by precedence.
	if parentHash == w.sectorHash {
		// clear all existing uploads
		q.clearUploads(w.id, 0)
	} else {
		var i int
		for i = 0; i < len(q.uploads[w.id]); i++ {
			if parentHash == q.uploads[w.id][i].hash {
				break
			}
		}

		if i == len(q.uploads[w.id]) {
			err = errors.New("upload has invalid parent hash")
			return
		}
		q.clearUploads(w.id, i)
	}

	uploadHash := SectorHash(newHashSet)
	u := upload{
		id: w.ID,
		requiredConfirmations: confirmations,
		hashSet:               newHashSet,
		hash:                  uploadHash,
		weight:                atomsChanged,
		deadline:              deadline,
	}

	weight = atomsChanged
	if s.uploads[w.ID] == nil {
		s.uploads[w.ID] = make([]*upload, 0)
	}
	q.uploads[w.ID] = append(q.uploads[w.ID], &u)
	q.updateWeight(w.ID, int(atomsChanged))
	q.insertEvent(&u)
	return
}*/
