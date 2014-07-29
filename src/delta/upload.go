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
	UploadID state.UploadID
	DeltaSet []Delta
}

func (e *Engine) UpdateSegment(sd SegmentDiff) (err error) {
	// Get the upload that corresponds with the upload id.
	upload, exists := e.state.ActiveUpload(sd.UploadID)
	if !exists {
		err = fmt.Errorf("Upload is not valid.")
		return
	}

	// See if the upload has already been completed.
	_, exists = e.completedUploads[sd.UploadID]
	if exists {
		err = fmt.Errorf("Upload already completed.")
		return
	}

	// Grab the wallet associated with the upload for future reference.
	var wallet state.Wallet
	wallet, err = e.state.LoadWallet(upload.WalletID)
	if err != nil {
		panic(err)
	}

	// See that the upload is active, which is done by verifying that the parent
	// hash either corresponds to the hash of the sector or has been marked as
	// completed in the engine.  See if the hash of the wallet matches the parent
	// hash of the upload.  If neither the hash of the wallet matches nor does
	// the hash of the parent exist in the list of completed uploads, then this
	// upload is not active - a preceeding upload needs to complete.
	_, exists = e.completedUploads[upload.ParentUploadID()]
	if upload.ParentHash != wallet.SectorSettings.Hash && !exists {
		err = fmt.Errorf("Upload is not active yet - please upload the data for the parent upload.")
		return
	}

	// Create the file that will store the upload, overwriting anything that's
	// potentially already there.
	filename := e.state.UploadFilename(upload)
	_, exists = e.completedUploads[upload.ParentUploadID()]
	if exists {
		// Need to copy the file from the parent sector modifier.
		parentUpload := e.completedUploads[upload.ParentUploadID()]
		err = siafiles.Copy(filename, e.state.UploadFilename(parentUpload))
	} else {
		// Copy the file from the wallet sector.
		err = siafiles.Copy(filename, e.state.SectorFilename(upload.WalletID))
	}

	// Open the duplicated file.
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// Apply the Deltas to the upload file.
	for _, delta := range sd.DeltaSet {
		if int(delta.Offset)+len(delta.Data) >= state.AtomSize*int(wallet.SectorSettings.Atoms) {
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

	// Run MerkleCollapse on this file, verify that the root hash is what the
	// upload indicates is correct.
	_, err = file.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	// root := state.MerkleCollapse(file)
	// if root != upload.HashSet[e.siblingIndex] {
	//	err = fmt.Errorf("Delta set did not result in the required merkle root.
	//	return
	// }

	// Set completedUploads for this upload
	e.completedUploads[upload.UploadID()] = upload

	// Upon returning, singal whether a completion should be loaded into the block.

	return
}
