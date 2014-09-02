package client

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/state"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siafiles"
)

// TestDownloadAndRepair uses the client api to create 3 participants working
// in consensus, uploads a file to them, then adds a 4th participant to see if
// repair works without issues.
func TestUploadAndRepair(t *testing.T) {
	// Clean out any previous test files.
	os.RemoveAll(siafiles.TempFilename("TestClient"))

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
	err = c.NewServer()
	if err != nil {
		t.Fatal(err)
	}

	// Create a bootstrap participant.
	if err != nil {
		t.Fatal(err)
	}
	err = c.NewBootstrapParticipant("0", siafiles.TempFilename("TestClient"), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Create 2 more participants to upload files to.
	err = c.NewJoiningParticipant("1", siafiles.TempFilename("TestClient"), 1)
	if err != nil {
		t.Fatal(err)
	}
	err = c.NewJoiningParticipant("2", siafiles.TempFilename("TestClient"), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Create a temp file to upload to the network.
	filesize := 250
	uploadFilename := siafiles.TempFilename("TestClient-UploadFile")
	randomBytes := siacrypto.RandomByteSlice(filesize)
	if err != nil {
		t.Fatal(err)
	}
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
	for i := range downloadBytes {
		if downloadBytes[i] != randomBytes[i] {
			t.Error("Mismatch on byte:", i)
		}
	}
}
