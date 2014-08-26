package client

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

func (c *Client) DownloadFile(id state.WalletID) (err error) {

}

// Submit a wallet request to the fountain wallet.
func (c *Client) RequestGenericWallet(id state.WalletID) (err error) {
	// Query to verify that the wallet id is available.
	var w state.Wallet
	err = c.router.SendMessage(network.Message{
		Dest: c.siblings[0].Address,
		Proc: "Participant.Wallet",
		Args: id,
		Resp: &w,
	})
	if err == nil {
		err = errors.New("Wallet already exists!")
		return
	}
	err = nil

	// Create a generic wallet with a keypair for the request.
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Fill out a keypair object and insert it into the generic wallet map.
	var kp Keypair
	kp.PublicKey = pk
	kp.SecretKey = sk

	// Get the current height of the quorum.
	currentHeight, err := c.GetHeight()
	if err != nil {
		return
	}

	// Send the requesting script input out to the network.
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: state.ScriptInput{
			WalletID: delta.FountainWalletID,
			Input:    delta.CreateFountainWalletInput(id, delta.DefaultScript(pk)),
			Deadline: currentHeight + state.MaxDeadline,
		},
		Resp: nil,
	})

	// Wait an appropriate amount of time for the request to be accepted: 2
	// blocks.
	time.Sleep(time.Duration(consensus.NumSteps) * 2 * consensus.StepDuration)

	// Query to verify that the request was accepted by the network.
	err = c.router.SendMessage(network.Message{
		Dest: c.siblings[0].Address,
		Proc: "Participant.Wallet",
		Args: id,
		Resp: &w,
	})
	if err != nil {
		return
	}
	if string(w.Script) != string(delta.DefaultScript(pk)) {
		err = errors.New("Wallet already exists - someone just beat you to it.")
		return
	}

	c.genericWallets[id] = kp

	return
}

// Send coins from one wallet to another.
func (c *Client) SendCoinGeneric(source state.WalletID, destination state.WalletID, amount state.Balance) (err error) {
	if _, ok := c.genericWallets[source]; !ok {
		err = errors.New("Could not access source wallet, perhaps it's not a generic wallet?")
		return
	}

	// Get the current height of the quorum, for setting the deadline on
	// the script input.
	currentHeight, err := c.GetHeight()
	if err != nil {
		return
	}

	input := state.ScriptInput{
		WalletID: source,
		Input:    delta.SendCoinInput(destination, amount),
		Deadline: currentHeight + state.MaxDeadline,
	}
	err = delta.SignScriptInput(&input, c.genericWallets[source].SecretKey)
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
	if _, exists := c.genericWallets[id]; !exists {
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

	// Figure out the number of atoms needed to upload the file.
	atomsNeeded := uint16(int(fileSize) / (int(state.AtomSize) * int(state.StandardK)))
	if int(fileSize)%(int(state.AtomSize)*int(state.StandardK)) != 0 {
		atomsNeeded++
	}
	if atomsNeeded > state.AtomsPerSector {
		err = errors.New("Cannot use such a large file.")
		return
	}

	// Create basic sector update.
	height, err := c.GetHeight()
	if err != nil {
		return
	}
	su := state.SectorUpdate{
		K: state.StandardK,
		ConfirmationsRequired: state.StandardConfirmations,
		Deadline:              height + 6,
	}

	// Get the set of erasure coded data to upload to the quorum.
	var segments [state.QuorumSize]*bytes.Buffer
	for i := range segments {
		segments[i] = new(bytes.Buffer)
	}
	var segmentsWriter [state.QuorumSize]io.Writer
	for i := range segments {
		segmentsWriter[i] = segments[i]
	}
	atoms, err := state.RSEncode(file, segmentsWriter, state.StandardK)
	if err != nil {
		return
	}
	su.Atoms = atoms

	// Now that we have written to the buffers, we have to convert them to
	// readers so they can be merkle hashed.
	var segmentBytes [state.QuorumSize][]byte
	var segmentReaders [state.QuorumSize]*bytes.Reader
	for i := range segments {
		segmentBytes[i] = segments[i].Bytes()
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
		Deadline: height + 4,
		Input:    delta.UpdateSectorInput(su),
		WalletID: id,
	}
	delta.SignScriptInput(&input, c.genericWallets[id].SecretKey)
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: input,
		Resp: nil,
	})

	// Wait 2 blocks while the update gets accepted.
	time.Sleep(consensus.StepDuration * time.Duration(state.QuorumSize))

	// Upload each segment to its respective sibling.
	for i := range segmentBytes {
		var accepted bool
		segmentUpload := delta.SegmentUpload{
			WalletID:    id,
			UpdateIndex: 0,
			NewSegment:  segmentBytes[i],
		}
		err = c.router.SendMessage(network.Message{
			Dest: c.siblings[i].Address,
			Proc: "Participant.UploadSegment",
			Args: segmentUpload,
			Resp: &accepted,
		})
	}

	return
}
