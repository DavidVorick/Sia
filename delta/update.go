package delta

import (
	"errors"
	"os"

	"github.com/NebulousLabs/Sia/state"
)

type SegmentUpload struct {
	WalletID    state.WalletID
	UpdateIndex uint32
	NewSegment  []byte
}

func (e *Engine) ProcessSegmentUpload(su SegmentUpload) (accepted bool, err error) {
	// Fetch the wallet and the update from the wallet.
	w, err := e.state.LoadWallet(su.WalletID)
	if err != nil {
		return
	}
	sectorUpdate, err := w.LoadSectorUpdate(su.UpdateIndex)
	if err != nil {
		return
	}

	// Check whether the update file has already been created, which
	// indicates whether the upload has already been completed for this
	// participant.
	filename := e.state.SectorUpdateFilename(su.WalletID, su.UpdateIndex)
	if _, err = os.Stat(filename); !os.IsNotExist(err) {
		err = errors.New("already have upload")
		return
	}

	// Check that the NewSegment will fit in the update.
	if len(su.NewSegment) > int(state.AtomSize)*int(sectorUpdate.Atoms) {
		err = errors.New("proposed update is larger than allocated update")
		return
	}

	// Create the file nad write the new segment.
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	_, err = file.Write(su.NewSegment)
	if err != nil {
		file.Close()
		os.Remove(filename)
		return
	}

	// Pad the remaining file with '0's
	numZerosNeeded := int(state.AtomSize)*int(sectorUpdate.Atoms) - len(su.NewSegment)
	empty := make([]byte, numZerosNeeded)
	_, err = file.Write(empty)
	if err != nil {
		file.Close()
		os.Remove(filename)
		return
	}

	// Run a merkle collapse on the file and see if the hash matches.
	_, err = file.Seek(0, 0)
	if err != nil {
		file.Close()
		os.Remove(filename)
		return
	}
	root, err := state.MerkleCollapse(file, sectorUpdate.Atoms)
	if err != nil {
		file.Close()
		os.Remove(filename)
		return
	}
	if root != sectorUpdate.HashSet[e.siblingIndex] {
		err = errors.New("hash does not match!")
		file.Close()
		os.Remove(filename)
		return
	}

	file.Close()
	accepted = true
	return
}

/*
import (
	"errors"
	"os"

	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

// TODO: add docstring
type Delta struct {
	Offset uint16
	Data   []byte
}

// TODO: add docstring
type SegmentDiff struct {
	UpdateID state.UpdateID
	DeltaSet []Delta
}

// TODO: add docstring
func (e *Engine) UpdateSegment(sd SegmentDiff) (accepted bool, err error) {
	// Fetch the update from the state, while verifying that it's a recognized
	// update.
	update, exists := e.state.GetSectorUpdate(sd.UpdateID)
	if !exists {
		err = errors.New("update is not valid")
		return
	}

	// See if the update has already been completed for this sibling. Completed
	// uploads is never cleaned out right now, meaning it's a map that's
	// continuously getting larger. I don't have a great solution.
	if e.completedUpdates[sd.UpdateID] {
		err = errors.New("update already completed")
		return
	}

	// Grab the wallet associated with the update to check the hash of the
	// sector being modified.
	var wallet state.Wallet
	wallet, err = e.state.LoadWallet(update.WalletID)
	if err != nil {
		panic(err)
	}

	// Verify that the update is active, which is done by verifying that the parent
	// hash either corresponds to the hash of the sector or has been marked as
	// completed in the engine.
	parentID := update.ParentID()
	parentCompleted := e.completedUpdates[parentID]
	if update.ParentCounter != wallet.SectorSettings.RecentUpdateCounter && !parentCompleted {
		err = errors.New("SectorUpdate is not active yet - please provide the updates for all parents")
		return
	}

	// Create the file that will house the non-commited update, overwriting
	// anything that's potentially already there. If the file is found in
	// completedUpdates, then the file must be created from a copy of the parent
	// update. Otherwise the file must be created from a copy of the wallet
	// segment.
	filename := e.state.UpdateFilename(sd.UpdateID)
	if e.completedUpdates[parentID] {
		err = siafiles.Copy(filename, e.state.UpdateFilename(parentID))
	} else {
		err = siafiles.Copy(filename, e.state.SectorFilename(update.WalletID))
	}

	// Open the duplicated file.
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// If the the number of atoms in the sector has been reduced, truncate the
	// file. If the number of atoms in the file has been increased, write all 0's
	// to the new space.
	var parentAtoms uint16
	// Need to account for the possibility that parent update is actually a wallet.
	parentUpdate, exists := e.state.GetSectorUpdate(parentID)
	if exists {
		parentAtoms = parentUpdate.Atoms
	} else {
		parentAtoms = wallet.SectorSettings.Atoms
	}
	if update.Atoms < parentAtoms {
		secErr := file.Truncate(int64(int(update.Atoms) * state.AtomSize))
		if secErr != nil {
			panic(secErr)
		}
	} else if update.Atoms > parentAtoms {
		// Write enough zeros at the end of the file such that it is the correct
		// segment size.
		zeroSlice := make([]byte, int(update.Atoms-parentAtoms)*state.AtomSize)
		_, secErr := file.WriteAt(zeroSlice, int64(int(parentAtoms)*state.AtomSize))
		if secErr != nil {
			panic(secErr)
		}
	}

	// Apply the deltas to the file.
	for _, delta := range sd.DeltaSet {
		if int(delta.Offset)+len(delta.Data) >= state.AtomSize*int(update.Atoms) {
			err = errors.New("invalid delta provided")
			return
		}

		_, err = file.WriteAt(delta.Data, int64(delta.Offset))
		if err != nil {
			panic(err)
		}
	}

	// Seek to the beginning of the file to run a MerkleCollapse.
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}

	// Run a MerkleCollapse on the file to verify that the deltas have produced
	// the right hash.
	root, err := state.MerkleCollapse(file, parentAtoms)
	if err != nil || root != update.HashSet[e.siblingIndex] {
		err = errors.New("delta set did not result in expected Merkle root")
		return
	}

	// Set completedUpdates for this upload.
	e.completedUpdates[sd.UpdateID] = true

	accepted = true
	return
}
*/
