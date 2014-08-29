package client

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

func (c *Client) genericWallet(gwid GenericWalletID) (gw *GenericWallet, err error) {
	gw, exists := c.genericWallets[gwid]
	if !exists {
		err = fmt.Errorf("could not find generic wallet of id %v", gwid)
		return
	}

	return
}

// Download the sector into filepath 'filename'.
func (gwid GenericWalletID) Download(c *Client, filename string) (err error) {
	// Get the wallet associated with the id.
	gw, err := c.genericWallet(gwid)
	if err != nil {
		return
	}

	// Download a segment from each sibling in the quorum, until StandardK
	// segments have been downloaded.
	var segments []io.Reader
	var indicies []byte
	for i := range c.metadata.Siblings {
		var segment []byte
		err = c.router.SendMessage(network.Message{
			Dest: c.metadata.Siblings[i].Address,
			Proc: "Participant.DownloadSegment",
			Args: gw.WalletID,
			Resp: &segment,
		})

		segments = append(segments, bytes.NewReader(segment))
		indicies = append(indicies, byte(i))
		if len(indicies) == state.StandardK {
			break
		}
	}

	// Check that enough pieces were retrieved.
	if len(indicies) < state.StandardK {
		err = errors.New("file not retrievable")
		return
	}

	// Create the file for writing.
	file, err := os.Create(filename)
	if err != nil {
		return
	}

	// Recover the StandardK segments into the file.
	_, err = state.RSRecover(segments, indicies, file, state.StandardK)
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
func (gwid GenericWalletID) SendCoin(c *Client, destination state.WalletID, amount state.Balance) (err error) {
	// Get the wallet associated with the id.
	gw, err := c.genericWallet(gwid)
	if err != nil {
		return
	}

	// Get the current height of the quorum, for setting the deadline on
	// the script input.
	input := state.ScriptInput{
		WalletID: gw.WalletID,
		Input:    delta.SendCoinInput(destination, amount),
		Deadline: c.metadata.Height + state.MaxDeadline,
	}
	err = delta.SignScriptInput(&input, gw.SecretKey)
	if err != nil {
		return
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})
	return
}

// Upload takes a file as input and uploads it to the wallet.
func (gwid GenericWalletID) Upload(c *Client, filename string) (err error) {
	// Get the wallet associated with the id.
	gw, err := c.genericWallet(gwid)
	if err != nil {
		return
	}

	// Refresh the metadata for greatest chance of success.
	c.RefreshMetadata()

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
		Deadline:              c.metadata.Height + 5,
	}

	// Create segments for the encoder output.
	var segments [state.QuorumSize]io.Writer
	var segmentsBuffer [state.QuorumSize]bytes.Buffer
	for i := range segments {
		segments[i] = &segmentsBuffer[i]
	}
	atoms, err := state.RSEncode(file, segments, state.StandardK)
	if err != nil {
		return
	}
	su.Atoms = atoms

	// Covert the buffer to byte slices containing the encoded data.
	var segmentBytes [state.QuorumSize][]byte
	for i := range segmentBytes {
		segmentBytes[i] = segmentsBuffer[i].Bytes()
	}

	// Now that we have written to the buffers, we have to convert them to
	// readers so they can be merkle hashed.
	var segmentReaders [state.QuorumSize]*bytes.Reader
	for i := range segments {
		segmentReaders[i] = bytes.NewReader(segmentBytes[i])
	}

	// Get the hashes of each segment.
	for i := range segmentReaders {
		su.HashSet[i], err = state.MerkleCollapse(segmentReaders[i], atoms)
		if err != nil {
			return
		}
	}

	// Submit the sector update.
	input := state.ScriptInput{
		Deadline: c.metadata.Height + 4,
		Input:    delta.UpdateSectorInput(su),
		WalletID: gw.WalletID,
	}
	delta.SignScriptInput(&input, c.genericWallets[gwid].SecretKey)
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})

	// Wait 3 blocks while the update gets accepted.
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize) * 3)

	// Upload each segment to its respective sibling.
	var successes byte
	for i := range segmentBytes {
		// Create a segment upload for the sibling of index 'i'.
		segmentUpload := delta.SegmentUpload{
			WalletID:    gw.WalletID,
			UpdateIndex: 0,
			NewSegment:  segmentBytes[i],
		}

		var accepted bool
		err2 := c.router.SendMessage(network.Message{
			Dest: c.metadata.Siblings[i].Address,
			Proc: "Participant.UploadSegment",
			Args: segmentUpload,
			Resp: &accepted,
		})
		if err2 == nil {
			successes++
		}
	}

	// Check that at least K segments were uploaded.
	if successes < state.StandardConfirmations {
		err = errors.New("not enough upload confirmations - upload failed")
		return
	}

	// Update the file size information attached to this wallet.
	gw.OriginalFileSize = fileSize

	return
}
