package state

import (
	"fmt"
	"siacrypto"
)

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

func (su *SectorUpdate) Hash() siacrypto.Hash {
	var hashSetBytes []byte
	for _, hash := range su.HashSet {
		hashSetBytes = append(hashSetBytes, hash[:]...)
	}
	return siacrypto.CalculateHash(hashSetBytes)
}

func (su *SectorUpdate) Expiration() uint32 {
	return su.EventExpiration
}

func (su *SectorUpdate) Counter() uint32 {
	return su.EventCounter
}

func (su *SectorUpdate) SetCounter(counter uint32) {
	su.EventCounter = counter
}

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
		// Logic to see if we have the file ourselves or not. If we do, simply copy
		// copy it over. If we don't, download it from the other guys.
	}

	// Call to delete the file that either did or did not exist.

	// Delete the completed uploads value within the engine..............
	// ah fudge this function is at the wrong level of abstraction.

	s.DeleteEvent(su)
}

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

func (s *State) InsertSectorUpdate(w *Wallet, su SectorUpdate) (err error) {
	// Check that the total atom usage of the wallet is not being overflowed.
	overflowCheck := uint32(w.CompensationWeight()) + uint32(su.Atoms)
	if overflowCheck > uint32(^uint16(0)) {
		err = fmt.Errorf("Adding the update atoms to the sector causes an overflow.")
		return
	}

	w.SectorSettings.UpdateAtoms += su.Atoms
	s.activeUpdates[su.WalletID] = append(s.activeUpdates[su.WalletID], su)
	return
}

func (su SectorUpdate) UpdateID() UpdateID {
	return UpdateID{
		WalletID: su.WalletID,
		Counter:  su.EventCounter,
	}
}

func (su SectorUpdate) ParentID() UpdateID {
	return UpdateID{
		WalletID: su.WalletID,
		Counter:  su.ParentCounter,
	}
}

func (s *State) UpdateFilename(id UpdateID) (filename string) {
	fmt.Sprintf("%s.sectorupdate.%s", s.walletFilename(id.WalletID), id.Counter)
	return
}

/* func (s *State) InsertSectorUpdate(su SectorUpdate) {
	// dont' forget to set the event counter on the update, though it might happen in InsertEvent
	s.InsertEvent(&u)
	s.AppendSectorModifier(&u)
	s.activeUploads[u.UploadID()] = &u
} */
