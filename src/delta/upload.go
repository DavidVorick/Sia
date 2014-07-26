package delta

import (
	"siacrypto"
	"state"
)

// An Upload is an event, which in particular means that it has an expiration
// associated with it. Upon expiring, it is checked whether the number of
// confirmations has reached the number of required confirmations. If yes, the
// HashSet of the upload becomes the HashSet of the sector, and the sector is
// successfully updated. If no, the upload is rejected and deleted from the
// system.
type Upload struct {
	// Which wallet is being modified.
	ID                    state.WalletID

	// The number of Siblings that need to receive the file diff before the
	// changes are accepted by the quorum.
	ConfirmationsRequired byte

	// An array of bools that indicate which siblings have confirmed that they
	// have received the upload.
	Confirmations         [state.QuorumSize]bool

	// The hash of the upload that is being changed. This hash must match the
	// most recent hash in the system, otherwise this upload is rejected as being
	// out-of-order or being out-of-date. This hash is required purely to prevent
	// synchronization problems. This hash is derived by appending all the hashes
	// in the HashSet into one set of state.QuorumSize * siacrypto.HashSize bytes and
	// hashing that.
	PreviousHash siacrypto.Hash

	// The MerkleCollapse value that each sibling should have after the segement
	// diff has been uploaded to them.
	HashSet               [state.QuorumSize]siacrypto.Hash

	// The number of atoms that are being altered during the diff. This is how
	// many atoms of temporary space will need to be allocated.
	AtomsAltered          uint16
}
