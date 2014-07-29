package delta

import (
	"fmt"
	"siacrypto"
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
*/

func (e *Engine) ProposeUpload(w *state.Wallet, confirmationsRequired byte, parentHash siacrypto.Hash, hashSet [state.QuorumSize]siacrypto.Hash, deadline uint32) (err error) {
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

	// Match the parent hash to the expected hash.
	if e.state.ActiveParentHash(*w, parentHash) {
		err = puerrNonCurrentParentHash
		return
	}

	// Verify that the quorum has enough atoms to support the upload. Long
	// term, this check won't be necessary because it'll be a part of the
	// preallocated resources planning.
	if e.state.Weight()+int(w.SectorSettings.Atoms) > int(state.AtomsPerQuorum) {
		err = puerrInsufficientAtoms
		return
	}

	// Update the wallet to reflect the new upload weight it has gained.
	w.SectorSettings.UploadAtoms += w.SectorSettings.Atoms

	// Update the eventlist to include an upload event.
	u := state.Upload{
		WalletID:              w.ID,
		ConfirmationsRequired: confirmationsRequired,
		ParentHash:            parentHash,
		HashSet:               hashSet,
	}
	e.state.InsertUpload(u)
	return
}
