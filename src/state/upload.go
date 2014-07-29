package state

import (
	"fmt"
	"siacrypto"
)

type UploadID [siacrypto.HashSize + WalletIDSize]byte

// An Upload is an event, which in particular means that it has an expiration
// associated with it. Upon expiring, it is checked whether the number of
// confirmations has reached the number of required confirmations. If yes, the
// HashSet of the upload becomes the HashSet of the sector, and the sector is
// successfully updated. If no, the upload is rejected and deleted from the
// system.
type Upload struct {
	// Which wallet is being modified.
	WalletID WalletID

	// The number of atoms that the sector will be resized to upon upload.
	NewAtomCount uint16

	// The MerkleCollapse value that each sibling should have after the segement
	// diff has been uploaded to them.
	HashSet [QuorumSize]siacrypto.Hash

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

	// Event variables.
	EventCounter    uint32
	EventExpiration uint32
}

func (u *Upload) WID() WalletID {
	return u.WalletID
}

func (u *Upload) Hash() siacrypto.Hash {
	var hashSetBytes []byte
	for _, hash := range u.HashSet {
		hashSetBytes = append(hashSetBytes, hash[:]...)
	}
	return siacrypto.CalculateHash(hashSetBytes)
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
	w, err := s.LoadWallet(u.WalletID)
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
		// Logic to see if we have the file ourselves or not. If we do, simply copy
		// copy it over. If we don't, download it from the other guys.
	}

	// Call to delete the file that either did or did not exist.

	// Delete the completed uploads value within the engine..............
	// ah fudge this function is at the wrong level of abstraction.

	s.DeleteEvent(u)
}

func (u Upload) UploadID() (uid UploadID) {
	hash := u.Hash()
	uidBytes := append(u.WalletID.Bytes(), hash[:]...)
	copy(uid[:], uidBytes)
	return
}

func (u Upload) ParentUploadID() (puid UploadID) {
	walletBytes := u.WalletID.Bytes()
	copy(puid[:], walletBytes)
	copy(puid[WalletIDSize:], u.ParentHash[:])
	return
}

func (s *State) UploadFilename(u Upload) (filename string) {
	hash := u.Hash()
	parentHash := string(hash[:])
	fmt.Sprintf("%s.upload.%s", s.walletFilename(u.WalletID), parentHash)
	return
}

func (s *State) InsertUpload(u Upload) {
	s.InsertEvent(&u)
	s.AppendSectorModifier(&u)
	s.activeUploads[u.UploadID()] = &u
}
