package delta

import (
	"errors"
	"siacrypto"
	"state"
)

// TODO: add docstring
// If these are really constants, they should be moved to instructions.go
const (
	CreateWalletCost = 8
	SendCost         = 6
	AddSiblingCost   = 500
)

var (
	errInsufficientBalance = errors.New("Insufficient balance to create a wallet with the given balance.")

	errNoEmptySiblings = errors.New("There are no empty spots in the quorum.")

	errUnallocatedSector      = errors.New("The sector has not been allocated, cannot make upload changes.")
	errTooManyConfirmations   = errors.New("Cannot require more than QuorumSize confirmations.")
	errTooFewConfirmations    = errors.New("Must require at least SectorSettings.K confirmations.")
	errNonCurrentParentHash   = errors.New("The parentHash given does not match the hash of the most recent upload to the quorum.")
	errDeadlineTooDistant     = errors.New("The deadline provided is more than state.MaxDeadline block into the future.")
	errDeadlineAlreadyExpired = errors.New("The deadline provided has already expired.")
	errAbsurdAtomsAltered     = errors.New("The number of atoms altered is greater than the number of atoms allocated.")
	errInsufficientAtoms      = errors.New("The quorum has insufficient atoms to support this upload.")
)

// CreateWallet takes an id, a Balance, and an initial script and uses
// those to create a new wallet that gets stored in stable memory.
// If a wallet of that id already exists then the process aborts.
func (e *Engine) CreateWallet(w *state.Wallet, childID state.WalletID, childBalance state.Balance, childScript []byte) (err error) {
	// Check that the wallet making the call has enough funds to deposit into the
	// wallet being created, and then subtract the funds from the parent wallet.
	if w.Balance.Compare(childBalance) < 0 {
		err = errInsufficientBalance
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

// AddSibling tries to add the new sibling to the existing quorum
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
		err = errNoEmptySiblings
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
	// w.sectorHash = siacrypto.HashBytes(firstAtom)

	return
}
*/

// TODO: add docstring
func (e *Engine) ProposeUpload(w *state.Wallet, confirmationsRequired byte, parentHash siacrypto.Hash, hashSet [state.QuorumSize]siacrypto.Hash, deadline uint32) (err error) {
	// Verify that the wallet in question has an allocated sector.
	if w.SectorSettings.Atoms < uint16(state.QuorumSize) {
		err = errUnallocatedSector
		return
	}

	// Verify that 'confirmationsRequired' is a legal value.
	if confirmationsRequired > state.QuorumSize {
		err = errTooManyConfirmations
		return
	} else if confirmationsRequired < w.SectorSettings.K {
		err = errTooFewConfirmations
		return
	}

	// Match the parent hash to the expected hash.
	if e.state.ActiveParentHash(*w, parentHash) {
		err = errNonCurrentParentHash
		return
	}

	// Verify that the quorum has enough atoms to support the upload. Long
	// term, this check won't be necessary because it'll be a part of the
	// preallocated resources planning.
	if e.state.Weight()+int(w.SectorSettings.Atoms) > int(state.AtomsPerQuorum) {
		err = errInsufficientAtoms
		return
	}

	// Update the wallet to reflect the new upload weight it has gained.
	w.SectorSettings.UploadAtoms += w.SectorSettings.Atoms

	// Update the eventlist to include an upload event.
	u := state.Upload{
		ID: w.ID,
		ConfirmationsRequired: confirmationsRequired,
		ParentHash:            parentHash,
		HashSet:               hashSet,
	}
	e.state.InsertEvent(&u)

	// Append the upload to the list of wallet sector modifiers.
	e.state.AppendSectorModifier(w.ID, &u)

	return
}
