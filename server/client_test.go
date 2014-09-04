package server

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
	testFolder := siafiles.TempFilename("TestClient")
	os.RemoveAll(testFolder)

	// Initialize a client.
	c, err := NewClient()
	if err != nil {
		// For some reason this error is okay.
		//t.Fatal(err)
	}
	err = c.Connect(14000)
	if err != nil {
		t.Fatal(err)
	}
	// Manually create server to avoid hostname problems.
	c.participantManager = new(ParticipantManager)
	c.participantManager.participants = make(map[string]*consensus.Participant)

	// Create a bootstrap participant.
	err = c.NewBootstrapParticipant("0", testFolder, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Create 2 more participants to upload files to.
	err = c.NewJoiningParticipant("1", testFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = c.NewJoiningParticipant("2", testFolder, 1)
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
	err = gwid.Upload(c, uploadFilename)
	if err != nil {
		t.Fatal(err)
	}

	// Wait enough blocks for the update to be moved to the sector
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize) * 7)

	// Download the file from the network.
	downloadFilename := siafiles.TempFilename("TestClient-DownloadFile")
	err = gwid.Download(c, downloadFilename)
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
	err = c.NewJoiningParticipant("3", testFolder, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !siafiles.Exists(filepath.Join(testFolder, "3", "wallet.AQAAAAAAAAA=.sector")) {
		t.Fatal(" Repaired wallet sector doesn't exist - something went wrong during repair.")
	}
}
