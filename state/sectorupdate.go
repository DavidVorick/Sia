package state

import (
	"errors"

	"github.com/NebulousLabs/Sia/siacrypto"
)

// TODO: add docstring
type UpdateID struct {
	WalletID WalletID
	Counter  uint32
}

// An Upload is an event, which in particular means that it has an expiration
// associated with it. Upon expiring, it is checked whether the number of
// confirmations has reached the number of required confirmations. If yes, the
// HashSet of the upload becomes the HashSet of the sector, and the sector is
// successfully updated. If no, the upload is rejected and deleted from the
// system.
type SectorUpdate struct {
	Index      uint32         // Update # for this wallet.
	ParentHash siacrypto.Hash // Hash of thi updates parent.

	// The updated SectorSettings values.
	Atoms uint16
	K     byte
	D     byte

	// The MerkleCollapse value that each sibling should have after the
	// segement diff has been uploaded to them. SectorSettings.Hash is the
	// hash of the HashSet.
	HashSet [QuorumSize]siacrypto.Hash

	// Confirmation variables. ConfirmationsRequired is the number of
	// confirmations needed before the update gets accepted to the network,
	// and Confirmatoins keeps track of which siblings have confirmed the
	// update.
	ConfirmationsRequired byte
	Confirmations         [QuorumSize]bool
}

// An update advancement is a tool that siblings use to signify that they have
// received the data necessary for them to complete their side of a sector
// update. When enough siblings have announced their update advancement, the
// update is rolled over into canon for the sector, and whatever siblings do
// not have the update will need to perform erasure recovery on the file, so
// they can get their piece.
type UpdateAdvancement struct {
	SiblingIndex byte
	Wallet       WalletID
	UpdateIndex  uint32
}

// AdvanceUpdate marks that the sibling of the particular index has signaled a
// completed update.
func (s *State) AdvanceUpdate(ua UpdateAdvancement) {
	/*
		for i := range s.activeUpdates[ua.UpdateID.WalletID] {
			if s.activeUpdates[ua.UpdateID.WalletID][i].EventCounter == ua.UpdateID.Counter {
				s.activeUpdates[ua.UpdateID.WalletID][i].Confirmations[ua.Index] = true
			}
		}
	*/
}

// Hash returns the hash of a SectorUpdate, which is really just a hash of the
// HashSet presented in the update.
func (su *SectorUpdate) Hash() siacrypto.Hash {
	fullSet := make([]byte, siacrypto.HashSize*int(QuorumSize))
	for i := range su.HashSet {
		copy(fullSet[i*siacrypto.HashSize:], su.HashSet[i][:])
	}
	return siacrypto.HashBytes(fullSet)
}

/*
// TODO: add docstring
func (su *SectorUpdate) HandleEvent(s *State) {
	// Load the wallet associated with the event.
	w, err := s.LoadWallet(su.WalletID)
	if err != nil {
		panic(err)
	}

	// Remove the weight on the wallet that the upload consumed.
	w.SectorSettings.UpdateAtoms -= w.SectorSettings.Atoms

	// Count the number of confirmations that the upload has received.
	var confirmationsReceived byte
	for _, confirmation := range su.Confirmations {
		if confirmation == true {
			confirmationsReceived += 1
		}
	}

	// If there are sufficient confirmations, update the sector hash values.
	if su.ConfirmationsRequired <= confirmationsReceived {
		// Hash our holding of the upload file and see if it matches the required
		// file. !! There are probably synchronization issues with doing things
		// this way.
		file, err := os.Open(s.UpdateFilename(su.UpdateID()))
		if err == nil {
			hash := MerkleCollapse(file)
			file.Close()
			if hash == su.HashSet[siblingIndex] {
				siafiles.Copy(s.SectorFilename(su.WalletID), s.UpdateFilename(su.UpdateID()))
			}
		}
	}

	// Call to delete the file that either did or did not exist.
	os.Remove(s.UpdateFilename(su.UpdateID()))

	s.DeleteEvent(su)
}
*/

// InsertSectorUpdate adds an update to a wallet, and to the event list.
func (s *State) InsertSectorUpdate(w *Wallet, su SectorUpdate) (err error) {
	// Check that the values in the update are legal.
	if su.Atoms > AtomsPerSector {
		err = errors.New("Sector allocates too many atoms")
		return
	}
	if su.K > MaxK {
		err = errors.New("Sector has a K exceeding the Max K")
		return
	}
	if su.K < MinK {
		err = errors.New("Sector has K below the Min K")
		return
	}
	if su.ConfirmationsRequired < MinConfirmations {
		err = errors.New("Confirmations required must be at least K!")
		return
	}

	// Check that there aren't already too many open updates.
	if len(w.SectorSettings.ActiveUpdates) >= MaxUpdates {
		err = errors.New("There are already the max number of open updates on the wallet")
		return
	}

	// Check that the hash of the most recent update matches the parent hash.
	if len(w.SectorSettings.ActiveUpdates) == 0 {
		if su.ParentHash != w.SectorSettings.Hash() {
			err = errors.New("Unrecognized parent hash - refusing to continue")
			return
		}
	} else {
		if su.ParentHash != w.SectorSettings.ActiveUpdates[len(w.SectorSettings.ActiveUpdates)-1].Hash() {
			err = errors.New("Parent hash doesn't match the most recent active upload.")
			return
		}
	}

	// Append the update to the list of active updates.
	w.SectorSettings.ActiveUpdates = append(w.SectorSettings.ActiveUpdates, su)

	// Create the event and put it into the event list.

	return
}
