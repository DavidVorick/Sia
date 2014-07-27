package delta

import (
	"fmt"
	"os"
	"siacrypto"
	"siaencoding"
	"state"
)

const (
	CreateWalletCost = 8
	SendCost         = 6
	AddSiblingCost   = 500
)

var (
	cwerrInsufficientBalance = fmt.Errorf("Insufficient balance to create a wallet with the given balance.")

	aserrNoEmptySiblings = fmt.Errorf("There are no empty spots in the quorum.")

	puerrUnallocatedSector      = fmt.Errorf("The sector has not been allocated, cannot make upload changes.")
	puerrTooManyConfirmations   = fmt.Errorf("Cannot require more than QuorumSize confirmations.")
	puerrTooFewConfirmations    = fmt.Errorf("Must require at least SectorSettings.K confirmations.")
	puerrNonCurrentParentHash   = fmt.Errorf("The parentHash given does not match the hash of the most recent upload to the quorum.")
	puerrDeadlineTooDistant     = fmt.Errorf("The deadline provided is more than state.MaxDeadline block into the future.")
	puerrDeadlineAlreadyExpired = fmt.Errorf("The deadline provided has already expired.")
	puerrAbsurdAtomsAltered     = fmt.Errorf("The number of atoms altered is greater than the number of atoms allocated.")
	puerrInsufficientAtoms      = fmt.Errorf("The quorum has insufficient atoms to support this upload.")
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
	if w.Balance.Compare(amount) < 0 {
		err = errors.New("insufficient balance")
		return
	}
	destWallet := s.LoadWallet(destID)
	if destWallet == nil {
		err = errors.New("destination wallet does not exist")
		return
	}

	w.Balance.Subtract(amount)
	destWallet.Balance.Add(amount)
	s.SaveWallet(destWallet)
	return
}

// Every wallet has a single sector, which can be up to 2^16 atoms of 4kb each,
// or 32GB total with 0 redundancy. Wallets pay for the size of their sector.
func (s *State) ResizeSectorErase(w *Wallet, atoms uint16, k byte) (cost int, weight int, err error) {
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
*/

func (e *Engine) ProposeUpload(w *state.Wallet, confirmationsRequired byte, parentHash siacrypto.Hash, hashSet [state.QuorumSize]siacrypto.Hash, atomsAltered uint16, deadline uint32) (err error) {
	// Verify that the wallet in question has an allocated sector.
	if w.SectorSettings.Atoms < uint16(state.QuorumSize) {
		err = puerrUnallocatedSector
		return
	}

	// Verify that 'confirmationsRequired' is a legal value.
	if confirmationsRequired > state.QuorumSize {
		err = puerrTooManyConfirmations
		return
	} else if confirmationsRequired < w.SectorSettings.K {
		err = puerrTooFewConfirmations
		return
	}

	// Verify that parentHash applies to this wallet.
	var expectedParentHash siacrypto.Hash
	activeUploads, exists := e.activeUploads[w.ID]
	if !exists {
		// Open the current wallet's segment and load the first QuorumSize atoms,
		// which contain the hash set for the current sector.
		sectorFilename := e.state.SectorFilename(w.ID)
		var file *os.File
		file, err = os.Open(sectorFilename)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Pull the parent hash set from the local segment.
		parentHashSet := make([]byte, int(state.QuorumSize)*siacrypto.HashSize)
		_, err = file.Read(parentHashSet)
		if err != nil {
			panic(err)
		}
		expectedParentHash = siacrypto.CalculateHash(parentHashSet)
	} else {
		// Calculate the hash of the most recently accepted active upload.
		mostRecentUpload := activeUploads[len(activeUploads)-1]
		var appendedHashSet []byte
		for _, hash := range mostRecentUpload.HashSet {
			appendedHashSet = append(appendedHashSet, hash[:]...)
		}
		expectedParentHash = siacrypto.CalculateHash(appendedHashSet)
	}

	// Match the parent hash to the expected hash.
	if expectedParentHash != parentHash {
		err = puerrNonCurrentParentHash
		return
	}

	// Verify that AtomsAltered makes sense.
	if atomsAltered > w.SectorSettings.Atoms {
		err = puerrAbsurdAtomsAltered
		return
	}

	// Verify that the quorum has enough atoms to support the upload. Long
	// term, this check won't be necessary because it'll be a part of the
	// preallocated resources planning.
	if e.state.Weight()+int(atomsAltered) > int(state.AtomsPerQuorum) {
		err = puerrInsufficientAtoms
		return
	}

	// Update the wallet to reflect the new upload weight it has gained.
	w.SectorSettings.UploadAtoms += atomsAltered

	// Update the eventlist to include an upload event.
	u := Upload{
		ID: w.ID,
		ConfirmationsRequired: confirmationsRequired,
		ParentHash:            parentHash,
		HashSet:               hashSet,
		AtomsAltered:          atomsAltered,
	}
	encodedUpload, err := siaencoding.Marshal(u)
	if err != nil {
		panic(err)
	}
	event := state.Event{
		Type:       "Upload",
		Expiration: deadline,
		// Counter will be set by InsertEvent.
		EncodedEvent: encodedUpload,
	}
	e.state.InsertEvent(event)

	return
}
