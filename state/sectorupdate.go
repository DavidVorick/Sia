package state

import (
	"errors"
	"os"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/siafiles"
)

const (
	StandardConfirmations = 3
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
	Index uint32 // Update # for this wallet.

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

	Deadline uint32
}

type SectorUpdateEvent struct {
	WalletID     WalletID
	UpdateIndex  uint32
	Deadline     uint32
	EventCounter uint32
}

// An update advancement is a tool that siblings use to signify that they have
// received the data necessary for them to complete their side of a sector
// update. When enough siblings have announced their update advancement, the
// update is rolled over into canon for the sector, and whatever siblings do
// not have the update will need to perform erasure recovery on the file, so
// they can get their piece.
type UpdateAdvancement struct {
	SiblingIndex byte
	WalletID     WalletID
	UpdateIndex  uint32
}

func (s *State) SectorUpdateFilename(id WalletID, index uint32) (filename string) {
	idBytes := siaencoding.EncUint64(uint64(id))
	idString := siafiles.SafeFilename(idBytes)
	indexBytes := siaencoding.EncUint32(index)
	indexString := siafiles.SafeFilename(indexBytes)
	filename = s.walletPrefix + "." + idString + ".update-" + indexString
	return
}

func (w *Wallet) LoadSectorUpdate(index uint32) (su SectorUpdate, err error) {
	for i := range w.SectorSettings.ActiveUpdates {
		if w.SectorSettings.ActiveUpdates[i].Index == index {
			su = w.SectorSettings.ActiveUpdates[i]
			return
		}
	}

	err = errors.New("could not find update of given index")
	return
}

func (sue *SectorUpdateEvent) Counter() uint32 {
	return sue.EventCounter
}

func (sue *SectorUpdateEvent) Expiration() uint32 {
	return sue.Deadline
}

func (sue *SectorUpdateEvent) HandleEvent(s *State) (err error) {
	// Need to be able to navigate from the event to the wallet.
	w, err := s.LoadWallet(sue.WalletID)
	if err != nil {
		return
	}

	su, err := w.LoadSectorUpdate(sue.UpdateIndex)
	if err != nil {
		return
	}

	// Remove the weight of the update from the wallet.
	w.SectorSettings.UpdateAtoms -= uint32(su.Atoms)

	// Count the number of confirmations.
	var confirmations int
	for _, confirmation := range su.Confirmations {
		if confirmation {
			confirmations++
		}
	}

	// Compare to the required confirmations.
	if confirmations >= int(su.ConfirmationsRequired) {
		// Remove all active updates leading to this update, inclusive.
		for i := range w.SectorSettings.ActiveUpdates {
			if i == int(sue.UpdateIndex) {
				w.SectorSettings.ActiveUpdates = w.SectorSettings.ActiveUpdates[i+1:]
				break
			}
		}

		w.SectorSettings.Atoms = su.Atoms
		w.SectorSettings.K = su.K
		w.SectorSettings.D = su.D
		w.SectorSettings.HashSet = su.HashSet

		// Copy the file from the update to the file for the sector.
		filename := s.SectorUpdateFilename(sue.WalletID, sue.UpdateIndex)
		if _, err = os.Stat(filename); os.IsNotExist(err) {
			// DO SOMETHING TO RECOVER THE FILE
		} else {
			siafiles.Copy(s.SectorFilename(sue.WalletID), filename)
			os.Remove(filename)
		}
	} else {
		// Remove all active updates following this update, inclusive.
		for i := range w.SectorSettings.ActiveUpdates {
			if w.SectorSettings.ActiveUpdates[i].Index == sue.UpdateIndex {
				w.SectorSettings.ActiveUpdates = w.SectorSettings.ActiveUpdates[:i]
				break
			}
		}
	}

	err = s.SaveWallet(w)
	if err != nil {
		return
	}

	return
}

func (sue *SectorUpdateEvent) SetCounter(newCounter uint32) {
	sue.EventCounter = newCounter
}

// AdvanceUpdate marks that the sibling of the particular index has signaled a
// completed update.
func (s *State) AdvanceUpdate(ua UpdateAdvancement) (err error) {
	w, err := s.LoadWallet(ua.WalletID)
	if err != nil {
		return
	}

	for i := range w.SectorSettings.ActiveUpdates {
		if w.SectorSettings.ActiveUpdates[i].Index == ua.UpdateIndex {
			w.SectorSettings.ActiveUpdates[i].Confirmations[ua.SiblingIndex] = true
			break
		}
	}

	err = s.SaveWallet(w)
	if err != nil {
		return
	}

	return
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

	// Check that the deadline is in bounds.
	if su.Deadline > s.Metadata.Height+MaxDeadline {
		err = errors.New("deadline too far in the future")
		return
	}

	// Figure out the update index.
	var index uint32
	if len(w.SectorSettings.ActiveUpdates) == 0 {
		index = 0
	} else {
		index = w.SectorSettings.ActiveUpdates[len(w.SectorSettings.ActiveUpdates)-1].Index + 1
	}

	// Append the update to the list of active updates.
	w.SectorSettings.ActiveUpdates = append(w.SectorSettings.ActiveUpdates, su)

	// Add the weight of the update to the wallet.
	w.SectorSettings.UpdateAtoms += uint32(su.Atoms)

	// Create the event and put it into the event list.
	sue := &SectorUpdateEvent{
		WalletID:    w.ID,
		UpdateIndex: index,
		Deadline:    su.Deadline,
	}
	s.InsertEvent(sue)

	return
}
