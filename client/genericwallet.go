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
	ID state.WalletID

	PublicKey siacrypto.PublicKey
	SecretKey siacrypto.SecretKey

	OriginalFileSize int64
}

// genericWallet is a helper function that fetches and returns a generic wallet
// when given a generic wallet id. It's not exported because the generic
// wallets are not meant to leave the client - all modifications that happen to
// them should be performed from within the client package.
func (c *Client) genericWallet(gwid GenericWalletID) (gw *GenericWallet, err error) {
	var exists bool
	*gw, exists = c.genericWallets[gwid]
	if !exists {
		err = fmt.Errorf("could not find generic wallet of id %v", gwid)
		return
	}

	return
}

// Download the sector into filepath 'filename'.
func (gw *GenericWallet) Download(c *Client, filename string) (err error) {
	// Download a segment from each sibling in the quorum, until StandardK
	// segments have been downloaded.
	var segments []io.Reader
	var indicies []byte
	for i := range c.metadata.Siblings {
		var segment []byte
		err = c.router.SendMessage(network.Message{
			Dest: c.metadata.Siblings[i].Address,
			Proc: "Participant.DownloadSegment",
			Args: gw.ID,
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
func (gw *GenericWallet) SendCoin(c *Client, destination state.WalletID, amount state.Balance) (err error) {
	// Get the current height of the quorum, for setting the deadline on
	// the script input.
	input := state.ScriptInput{
		WalletID: gw.ID,
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

// UpdateSectorGeneric takes the id of a generic wallet, along with a file, and
// replaces whatever sector/file is currently housed in the generic wallet with
// the new file.
func (c *Client) UploadFile(id state.WalletID, filename string) (err error) {
	// Check that the wallet is available to this client.
	if _, exists := c.genericWallets[GenericWalletID(id)]; !exists {
		err = errors.New("do not have access to given wallet")
		return
	}

	// Get a fresh list of siblings.
	c.RefreshSiblings()

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
		Deadline:              c.metadata.Height + 6,
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
		WalletID: id,
	}
	delta.SignScriptInput(&input, c.genericWallets[GenericWalletID(id)].SecretKey)
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})

	// Wait 3 blocks while the update gets accepted.
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize) * 3)

	// Upload each segment to its respective sibling.
	for i := range segmentBytes {
		var accepted bool
		segmentUpload := delta.SegmentUpload{
			WalletID:    id,
			UpdateIndex: 0,
			NewSegment:  segmentBytes[i],
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.metadata.Siblings[i].Address,
			Proc: "Participant.UploadSegment",
			Args: segmentUpload,
			Resp: &accepted,
		})

		if err != nil {
			println("error, but handle bad.")
		}
	}

	originalKeypair := c.genericWallets[GenericWalletID(id)]
	originalKeypair.OriginalFileSize = fileSize
	c.genericWallets[GenericWalletID(id)] = originalKeypair

	return
}
