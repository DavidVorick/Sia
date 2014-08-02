package state

import (
	"siacrypto"
)

// An Upload is an event, which in particular means that it has an expiration
// associated with it. Upon expiring, it is checked whether the number of
// confirmations has reached the number of required confirmations. If yes, the
// HashSet of the upload becomes the HashSet of the sector, and the sector is
// successfully updated. If no, the upload is rejected and deleted from the
// system.
type Upload struct {
	// Which wallet is being modified.
	ID WalletID

	// The number of Siblings that need to receive the file diff before the
	// changes are accepted by the quorum.
	ConfirmationsRequired byte

	// An array of bools that indicate which siblings have confirmed that they
	// have received the upload.
	Confirmations [QuorumSize]bool

	// The hash of the upload that is being changed. This hash must match the
	// most recent hash in the system, otherwise this upload is rejected as being
	// out-of-order or being out-of-date. This hash is required purely to prevent
	// synchronization problems. This hash is derived by appending all the hashes
	// in the HashSet into one set of QuorumSize * siacrypto.HashSize bytes and
	// hashing that.
	ParentHash siacrypto.Hash

	// The MerkleCollapse value that each sibling should have after the segement
	// diff has been uploaded to them.
	HashSet [QuorumSize]siacrypto.Hash

	// Event variables.
	EventCounter    uint32
	EventExpiration uint32
}

func (u *Upload) Hash() siacrypto.Hash {
	var hashSetBytes []byte
	for _, hash := range u.HashSet {
		hashSetBytes = append(hashSetBytes, hash[:]...)
	}
	return siacrypto.HashBytes(hashSetBytes)
}

func (u *Upload) Expiration() uint32 {
	return u.EventExpiration
}

func (u *Upload) Counter() uint32 {
	return u.EventCounter
}

func (u *Upload) SetCounter(counter uint32) {
	u.EventCounter = counter
}

func (u *Upload) HandleEvent(s *State) {
	// Load the wallet associated with the event.
	w, err := s.LoadWallet(u.ID)
	if err != nil {
		panic(err)
	}

	// Remove the weight on the wallet that the upload consumed.
	w.SectorSettings.UploadAtoms -= w.SectorSettings.Atoms

	// Count the number of confirmations that the upload has received.
	var confirmationsReceived byte
	for _, confirmation := range u.Confirmations {
		if confirmation == true {
			confirmationsReceived += 1
		}
	}

	// If there are sufficient confirmations, update the sector hash values.
	if u.ConfirmationsRequired <= confirmationsReceived {
		file, err := s.OpenUpload(u.ID, u.ParentHash)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	s.DeleteEvent(u)
}
