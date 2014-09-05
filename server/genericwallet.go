package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// A GenericWalletID is currently just a normal WalletID, but enforces type
// safety when making calls to generic wallet functions in the client. Also,
// when asked for a generic wallet, the client will only return the ID instead
// of the whole wallet; this is a safer way to interact with the wallet type.
type GenericWalletID state.WalletID

// A GenericWallet is a transportable struct which points to a wallet in the
// quorum of id 'id'. The wallet is assumed to have a generic script body which
// uses 'PublicKey' as the public key. The data stored in the wallet's sector
// is assumed to have a 'K' value of 'state.StandardK', and is assumed to
// decode to a file exactly 'OriginalFileSize' bytes on disk.
type GenericWallet struct {
	WalletID state.WalletID

	PublicKey siacrypto.PublicKey
	SecretKey siacrypto.SecretKey

	OriginalFileSize int64
}

// Helper function on GenericWallet that calculates and returns the id to the
// wallet.
func (gw GenericWallet) ID() GenericWalletID {
	return GenericWalletID(gw.WalletID)
}

// Returns the generic wallet associated with the wallet id. This is an
// exported function, and to protect the internal state of the client, only a
// copy of the generic wallet is returned.
func (s *Server) GenericWallet(gwid GenericWalletID) (gw GenericWallet, err error) {
	gwPointer, exists := s.genericWallets[gwid]
	if !exists {
		err = fmt.Errorf("could not find generic wallet of id %v", gwid)
		return
	}
	gw = *gwPointer

	return
}

// Returns the generic wallet assiciated with the wallet id. This is a
// non-exported function, and return a pointer to the generic wallet stored
// within the client. Editing the wallet returned by this function will modify
// the clients internal state.
func (s *Server) genericWallet(gwid GenericWalletID) (gw *GenericWallet, err error) {
	gw, exists := s.genericWallets[gwid]
	if !exists {
		err = fmt.Errorf("could not find generic wallet of id %v", gwid)
		return
	}

	return
}

// Download the sector into filepath 'filename'.
func (gwid GenericWalletID) Download(s *Server, filename string) (err error) {
	// Get the wallet associated with the id.
	gw, err := s.genericWallet(gwid)
	if err != nil {
		return
	}

	// Download a segment from each sibling in the quorum, until StandardK
	// segments have been downloaded.
	var segments []io.Reader
	var indices []byte
	for i := range s.metadata.Siblings {
		var segment []byte
		err2 := s.router.SendMessage(network.Message{
			Dest: s.metadata.Siblings[i].Address,
			Proc: "Participant.DownloadSegment",
			Args: gw.WalletID,
			Resp: &segment,
		})
		if err2 != nil {
			continue
		}

		segments = append(segments, bytes.NewReader(segment))
		indices = append(indices, byte(i))
		if len(indices) == state.StandardK {
			break
		}
	}

	// Check that enough pieces were retrieved.
	if len(indices) < state.StandardK {
		err = errors.New("file not retrievable")
		return
	}

	// Create the file for writing.
	file, err := os.Create(filename)
	if err != nil {
		return
	}

	// Recover the StandardK segments into the file.
	_, err = state.RSRecover(segments, indices, file, state.StandardK)
	if err != nil {
		return
	}

	// Truncate the file to it's original size, effectively removing any
	// padding that may have been added.
	err = file.Truncate(gw.OriginalFileSize)
	if err != nil {
		return
	}

	return
}

// Send coins to wallet 'destination'.
func (gwid GenericWalletID) SendCoin(s *Server, destination state.WalletID, amount state.Balance) (err error) {
	// Get the wallet associated with the id.
	gw, err := s.genericWallet(gwid)
	if err != nil {
		return
	}

	// Get the current height of the quorum, for setting the deadline on
	// the script input.
	input := state.ScriptInput{
		WalletID: gw.WalletID,
		Input:    delta.SendCoinInput(destination, amount),
		Deadline: s.metadata.Height + state.MaxDeadline,
	}
	err = delta.SignScriptInput(&input, gw.SecretKey)
	if err != nil {
		return
	}

	s.broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})
	return
}

// Upload takes a file as input and uploads it to the wallet.
func (gwid GenericWalletID) Upload(s *Server, filename string) (err error) {
	// Get the wallet associated with the id.
	gw, err := s.genericWallet(gwid)
	if err != nil {
		return
	}

	// Refresh the metadata for greatest chance of success.
	s.refreshMetadata()

	// Calculate the size of the file.
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return
	}
	fileSize := info.Size()

	// Create basic sector update.
	su := state.SectorUpdate{
		K: state.StandardK,
		ConfirmationsRequired: state.StandardConfirmations,
	}
	su.Event.Deadline = s.metadata.Height + 5

	// Create segments for the encoder output.
	segments := make([][]byte, state.QuorumSize)
	atoms, err := state.RSEncode(file, segments, state.StandardK)
	if err != nil {
		return
	}
	su.Atoms = atoms

	// Get the hashes of each segment.
	for i := range segments {
		su.HashSet[i], err = state.MerkleCollapse(bytes.NewReader(segments[i]), atoms)
		if err != nil {
			return
		}
	}

	// Submit the sector update.
	input := state.ScriptInput{
		Deadline: s.metadata.Height + 4,
		Input:    delta.UpdateSectorInput(su),
		WalletID: gw.WalletID,
	}
	delta.SignScriptInput(&input, gw.SecretKey)
	s.broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})

	// Wait 3 blocks while the update gets accepted.
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize) * 3)

	// Upload each segment to its respective sibling.
	var successes byte
	for i := range segments {
		// Create a segment upload for the sibling of index 'i'.
		segmentUpload := delta.SegmentUpload{
			WalletID:    gw.WalletID,
			UpdateIndex: 0,
			NewSegment:  segments[i],
		}

		var accepted bool
		sendErr := s.router.SendMessage(network.Message{
			Dest: s.metadata.Siblings[i].Address,
			Proc: "Participant.UploadSegment",
			Args: segmentUpload,
			Resp: &accepted,
		})
		if sendErr == nil {
			successes++
		}
	}

	// Check that at least K segments were uploaded.
	if successes < state.StandardConfirmations {
		err = fmt.Errorf("not enough upload confirmations - need %v, got %v", state.StandardConfirmations, successes)
		return
	}

	// Update the file size information attached to this wallet.
	gw.OriginalFileSize = fileSize

	return
}
