package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
	"github.com/NebulousLabs/Sia/state"
)

// TestDownloadAndRepair uses the client api to create 3 participants working
// in consensus, uploads a file to them, then adds a 4th participant to see if
// repair works without issues.
func TestUploadAndRepair(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Clean out any previous test files.
	config.Filesystem.ParticipantDir = siafiles.TempFilename("TestClient")
	os.RemoveAll(config.Filesystem.ParticipantDir)

	// Initialize a client.
	s := newServer()
	err := s.connect(14000, false)
	if err != nil {
		// a 'local only' error gets returned, which is intentional.
		// t.Fatal(err)
	}
	// Manually create server to avoid hostname problems.
	s.participantManager = new(ParticipantManager)
	s.participantManager.participants = make(map[string]*consensus.Participant)

	// Create a bootstrap participant.
	npi := NewParticipantInfo{
		Name:      "0",
		SiblingID: 1,
	}

	err = s.NewBootstrapParticipant(npi, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create 2 more participants to upload files to.
	npi.Name = "1"
	err = s.NewJoiningParticipant(npi, nil)
	if err != nil {
		t.Fatal(err)
	}
	npi.Name = "2"
	err = s.NewJoiningParticipant(npi, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a temp file to upload to the network.
	filesize := 250
	uploadFilename := siafiles.TempFilename("TestClient-UploadFile")
	randomBytes := siacrypto.RandomByteSlice(filesize)
	file, err := os.Create(uploadFilename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write(randomBytes)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Upload a file to the network.
	gwid := GenericWalletID(1)
	gup := GenericUploadParams{
		GWID:     gwid,
		Filename: uploadFilename,
	}
	err = s.GenericUpload(gup, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Wait enough blocks for the update to be moved to the sector
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize) * 7)

	// Download the file from the network.
	downloadFilename := siafiles.TempFilename("TestClient-DownloadFile")
	gdp := GenericDownloadParams{
		GWID:     gwid,
		Filename: downloadFilename,
	}
	err = s.GenericDownload(gdp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the downloaded file is identical to the original.
	downloadBytes, err := ioutil.ReadFile(downloadFilename)
	if err != nil {
		t.Fatal(err)
	}
	if len(downloadBytes) != len(randomBytes) {
		t.Error("Mismatch on file lengths between download and original.")
	} else {
		for i := range downloadBytes {
			if downloadBytes[i] != randomBytes[i] {
				t.Error("Mismatch on byte:", i)
			}
		}
	}

	// Add another participant and see that the repair triggers.
	npi.Name = "3"
	err = s.NewJoiningParticipant(npi, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !siafiles.Exists(filepath.Join(config.Filesystem.ParticipantDir, "3", "wallet.AQAAAAAAAAA=.sector")) {
		t.Fatal(" Repaired wallet sector doesn't exist - something went wrong during repair.")
	}
}
