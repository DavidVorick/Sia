package delta

import (
	"fmt"
	"os"
	"siafiles"
	"state"
)

type Delta struct {
	Offset uint16
	Data   []byte
}

type SegmentDiff struct {
	UpdateID state.UpdateID
	DeltaSet []Delta
}

func (e *Engine) UpdateSegment(sd SegmentDiff) (accepted bool, err error) {
	// Fetch the update from the state, while verifying that it's a recognized
	// update.
	update, exists := e.state.GetSectorUpdate(sd.UpdateID)
	if !exists {
		err = fmt.Errorf("Update is not valid.")
		return
	}

	// See if the update has already been completed for this sibling. Completed
	// uploads is never cleaned out right now, meaning it's a map that's
	// continuously getting larger. I don't have a great solution.
	if e.completedUpdates[sd.UpdateID] {
		err = fmt.Errorf("Update already completed.")
		return
	}

	// Grab the wallet associated with the update to check the hash of the sector
	// being modified.
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
		err = fmt.Errorf("SectorUpdate is not active yet - please provide the updates for all parents.")
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
	parentUpdate, exists := e.state.GetSectorUpdate(parentID) // Need to account for the possibility that parent update is actually a wallet.
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
		// Seek to the end of the file.
		_, secErr := file.Seek(int64(int(parentAtoms)*state.AtomSize), 0)
		if secErr != nil {
			panic(secErr)
		}

		// Write enough zeros at the end of the file such that it is the correct
		// segment size.
		zeroSlice := make([]byte, int(update.Atoms-parentAtoms)*state.AtomSize)
		_, secErr = file.Write(zeroSlice)
		if secErr != nil {
			panic(secErr)
		}
	}

	// Apply the deltas to the file.
	for _, delta := range sd.DeltaSet {
		if int(delta.Offset)+len(delta.Data) >= state.AtomSize*int(update.Atoms) {
			err = fmt.Errorf("Invalid delta provided")
			return
		}

		_, err = file.Seek(int64(delta.Offset), 0)
		if err != nil {
			panic(err)
		}

		_, err = file.Write(delta.Data)
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
		err = fmt.Errorf("Delta set did not result in the required merkle root.")
		return
	}

	// Set completedUpdates for this upload.
	e.completedUpdates[sd.UpdateID] = true

	// Set accepted to true, signaling that the update has been accepted.
	accepted = true
	return
}
