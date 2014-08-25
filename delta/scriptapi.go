package delta

import (
	"errors"

	"github.com/NebulousLabs/Sia/state"
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

	errInvalidK             = errors.New("K must hold either a value of 1 or 2.")
	errTooFewAtoms          = errors.New("A sector must have more than QuorumSize atoms.")
	errUnallocatedSector    = errors.New("The sector has not been allocated, cannot make upload changes.")
	errTooManyConfirmations = errors.New("Cannot require more than QuorumSize confirmations.")
	errTooFewConfirmations  = errors.New("Must require at least SectorSettings.K confirmations.")
	errNonCurrentParentID   = errors.New("The parentHash given does not match the hash of the most recent upload to the quorum.")
	errAbsurdAtomsAltered   = errors.New("The number of atoms altered is greater than the number of atoms allocated.")
	errInsufficientAtoms    = errors.New("The quorum has insufficient atoms to support this upload.")
	errDeadlineTooEarly     = errors.New("An upload takes at least 2 complete blocks to succeed - deadline too early.")
	errLongDeadline         = errors.New("The deadline is too far in the future.")
)

// AddSibling tries to add the new sibling to the existing quorum
// and throws the sibling out if there's no space. Once quorums are
// communicating, the AddSibling routine will always succeed.
func (e *Engine) AddSibling(w *state.Wallet, sib state.Sibling) (err error) {
	// Down payment stuff will be added here.

	// Look through the quorum for an empty sibling.
	for i := byte(0); i < state.QuorumSize; i++ {
		if e.state.Metadata.Siblings[i].Inactive() {
			sib.Status = state.SiblingPassiveWindow
			sib.Index = i
			sib.WalletID = w.ID
			e.state.Metadata.Siblings[i] = sib
			break
		}
	}

	if sib.Inactive() {
		err = errNoEmptySiblings
		return
	}

	// Charge the wallet some volume that's required as a down payment.

	return
}

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

	// Insert the child wallet.
	err = e.state.InsertWallet(childWallet)
	if err != nil {
		return
	}

	return
}

// Send is a call that sends siacoins from the source wallet to the destination
// wallet.
func (e *Engine) SendCoin(w *state.Wallet, amount state.Balance, destID state.WalletID) (err error) {
	// Check that the source wallet contains enough to send the desired
	// amount.
	if w.Balance.Compare(amount) < 0 {
		err = errors.New("insufficient balance")
		return
	}

	// Check that the destination wallet is available.
	destWallet, err := e.state.LoadWallet(destID)
	if err != nil {
		return
	}

	// Commit the send.
	w.Balance.Subtract(amount)
	destWallet.Balance.Add(amount)
	e.state.SaveWallet(destWallet)

	return
}

// UpdateSector takes a different approach, which is essentially to completely
// outsource the function to the state package, reporting an error if needed. I
// have no idea if this is a good approach, but at somepoint we'll need to
// standardize around a single approach. It's probably better that this file is
// the leaner.
func (e *Engine) UpdateSector(w *state.Wallet, su state.SectorUpdate) (err error) {
	err = e.state.InsertSectorUpdate(w, su)
	if err != nil {
		return
	}

	return
}
