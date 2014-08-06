package state

import (
	"fmt"
	//"os"

	"github.com/NebulousLabs/Sia/siacrypto"
	//"github.com/NebulousLabs/Sia/siafiles"
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
	// Which wallet is being modified.
	WalletID WalletID

	// The hash of the upload that is being changed. This hash must match the
	// most recent hash in the system, otherwise this upload is rejected as being
	// out-of-order or being out-of-date. This hash is required purely to prevent
	// synchronization problems. This hash is derived by appending all the hashes
	// in the HashSet into one set of QuorumSize * siacrypto.HashSize bytes and
	// hashing that.
	ParentCounter uint32

	// The updated SectorSettings value.
	Atoms uint16
	K     byte
	D     byte

	// The MerkleCollapse value that each sibling should have after the segement
	// diff has been uploaded to them. SectorSettings.Hash is the hash of the
	// HashSet.
	HashSet [QuorumSize]siacrypto.Hash

	// Confirmation variables. ConfirmationsRequired is the number of
	// confirmations needed before the update gets accepted to the network, and
	// Confirmatoins keeps track of which siblings have confirmed the update.
	ConfirmationsRequired byte
	Confirmations         [QuorumSize]bool

	// Event variables, as UpdateSector is an event.
	EventCounter    uint32
	EventExpiration uint32
}

// TODO: add docstring
type UpdateAdvancement struct {
	Index    byte
	UpdateID UpdateID
}

// Hash returns the hash of a SectorUpdate.
func (su *SectorUpdate) Hash() siacrypto.Hash {
	var hashSetBytes []byte
	for _, hash := range su.HashSet {
		hashSetBytes = append(hashSetBytes, hash[:]...)
	}
	return siacrypto.HashBytes(hashSetBytes)
}

// Expiration is a getter that returns the EventExpiration of the SectorUpdate.
func (su *SectorUpdate) Expiration() uint32 {
	return su.EventExpiration
}

// Expiration is a getter that returns the EventCounter of the SectorUpdate.
func (su *SectorUpdate) Counter() uint32 {
	return su.EventCounter
}

// Expiration is a setter that sets the EventCounter of the SectorUpdate.
func (su *SectorUpdate) SetCounter(counter uint32) {
	su.EventCounter = counter
}

// TODO: add docstring
func (su *SectorUpdate) HandleEvent(s *State) {
	/*
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
	*/
}

// TODO: add docstring
func (s *State) AvailableParentID(parentID UpdateID) bool {
	// If there is an entry for this wallet in the active updates map, then
	// that's the only thing that counts. Otherwise, the value of the wallet is
	// the only thing that counts.
	updateList, exists := s.activeUpdates[parentID.WalletID]
	if exists {
		// Compare the counter in the parent to the counter of the latest element
		// in the update list.
		if parentID.Counter == updateList[len(updateList)-1].EventCounter {
			return true
		}
	} else {
		// Compare the counter in the parent to the counter of the wallet.
		w, err := s.LoadWallet(parentID.WalletID)
		if err != nil {
			return false
		}

		if parentID.Counter == w.SectorSettings.RecentUpdateCounter {
			return true
		}
	}
	return false
}

// TODO: add docstring
func (s *State) GetSectorUpdate(uid UpdateID) (update SectorUpdate, exists bool) {
	updateList, exists := s.activeUpdates[uid.WalletID]
	if !exists {
		return
	}

	for _, update = range updateList {
		if update.EventCounter == uid.Counter {
			return
		}
	}
	exists = false
	return
}

// TODO: add docstring
func (s *State) InsertSectorUpdate(w *Wallet, su SectorUpdate) (err error) {
	// Check that the total atom usage of the wallet is not being overflowed.
	overflowCheck := uint32(w.CompensationWeight()) + uint32(su.Atoms)
	if overflowCheck > uint32(^uint16(0)) {
		err = fmt.Errorf("Adding the update atoms to the sector causes an overflow.")
		return
	}

	s.InsertEvent(&su)
	w.SectorSettings.UpdateAtoms += su.Atoms
	s.activeUpdates[su.WalletID] = append(s.activeUpdates[su.WalletID], su)
	return
}

// TODO: add docstring
func (su SectorUpdate) UpdateID() UpdateID {
	return UpdateID{
		WalletID: su.WalletID,
		Counter:  su.EventCounter,
	}
}

// TODO: add docstring
func (su SectorUpdate) ParentID() UpdateID {
	return UpdateID{
		WalletID: su.WalletID,
		Counter:  su.ParentCounter,
	}
}

// UpdateFilename creates a filename corresponding to a given UpdateID.
func (s *State) UpdateFilename(id UpdateID) (filename string) {
	return fmt.Sprintf("%s.sectorupdate.%d", s.walletFilename(id.WalletID), id.Counter)
}

// AdvanceUpdate marks that the sibling of the particular index has signaled a
// completed update.
func (s *State) AdvanceUpdate(ua UpdateAdvancement) {
	for i := range s.activeUpdates[ua.UpdateID.WalletID] {
		if s.activeUpdates[ua.UpdateID.WalletID][i].EventCounter == ua.UpdateID.Counter {
			s.activeUpdates[ua.UpdateID.WalletID][i].Confirmations[ua.Index] = true
		}
	}
}
