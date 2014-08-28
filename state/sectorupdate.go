package state

import (
	"errors"

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

	// The updated Sector values.
	Atoms uint16
	K     byte
	D     byte

	// The MerkleCollapse value that each sibling should have after the
	// segement diff has been uploaded to them. Sector.Hash is the
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
	for i := range w.Sector.ActiveUpdates {
		if w.Sector.ActiveUpdates[i].Index == index {
			su = w.Sector.ActiveUpdates[i]
			return
		}
	}

	err = errors.New("could not find update of given index")
	return
}

// AdvanceUpdate marks that the sibling of the particular index has signaled a
// completed update.
func (s *State) AdvanceUpdate(ua UpdateAdvancement) (err error) {
	w, err := s.LoadWallet(ua.WalletID)
	if err != nil {
		return
	}

	for i := range w.Sector.ActiveUpdates {
		if w.Sector.ActiveUpdates[i].Index == ua.UpdateIndex {
			w.Sector.ActiveUpdates[i].Confirmations[ua.SiblingIndex] = true
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
	if len(w.Sector.ActiveUpdates) >= MaxUpdates {
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
	if len(w.Sector.ActiveUpdates) == 0 {
		index = 0
	} else {
		index = w.Sector.ActiveUpdates[len(w.Sector.ActiveUpdates)-1].Index + 1
	}

	// Append the update to the list of active updates.
	w.Sector.ActiveUpdates = append(w.Sector.ActiveUpdates, su)

	// Add the weight of the update to the wallet.
	w.Sector.UpdateAtoms += uint32(su.Atoms)

	// Create the event and put it into the event list.
	sue := &SectorUpdateEvent{
		WalletID:    w.ID,
		UpdateIndex: index,
		Deadline:    su.Deadline,
	}
	s.InsertEvent(sue, true)

	return
}
