package state

import (
	"os"

	"github.com/NebulousLabs/Sia/siafiles"
)

type SectorUpdateEvent struct {
	WalletID     WalletID
	UpdateIndex  uint32
	Deadline     uint32
	EventCounter uint32
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
