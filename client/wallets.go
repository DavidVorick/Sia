package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/siaencoding"
	"github.com/NebulousLabs/Sia/state"
)

// Returns a list of all wallets available to the client.
func (c *Client) GetWalletIDs() (ids []state.WalletID) {
	ids = make([]state.WalletID, 0, len(c.genericWallets))
	for id := range c.genericWallets {
		ids = append(ids, id)
	}
	// Add other types of wallets as they are implemented.
	return
}

// Wallet type takes an id as input and returns the wallet type. An error is
// returned if the wallet is not found by the client.
func (c *Client) WalletType(id state.WalletID) (walletType string, err error) {
	// Check if the wallet is a generic type.
	_, exists := c.genericWallets[id]
	if exists {
		walletType = "generic"
		return
	}

	// Check for other types of wallets.

	err = errors.New("Wallet is not available.")
	return
}

func (c *Client) GenericWallet(id state.WalletID) (gw *GenericWallet, err error) {
	if _, exists := c.genericWallets[id]; exists {
		*gw = c.genericWallets[id]
	} else {
		err = fmt.Errorf("could not find generic wallet of id %v", id)
	}

	return
}

// Submit a wallet request to the fountain wallet.
func (c *Client) RequestGenericWallet(id state.WalletID) (err error) {
	// Query to verify that the wallet id is available.
	var w state.Wallet
	err = c.router.SendMessage(network.Message{
		Dest: c.metadata.Siblings[0].Address,
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
	var gw GenericWallet
	gw.PublicKey = pk
	gw.SecretKey = sk

	// Get the current height of the quorum.
	// Send the requesting script input out to the network.
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: state.ScriptInput{
			WalletID: delta.FountainWalletID,
			Input:    delta.CreateFountainWalletInput(id, delta.DefaultScript(pk)),
			Deadline: c.metadata.Height + state.MaxDeadline,
		},
		Resp: nil,
	})

	// Wait an appropriate amount of time for the request to be accepted: 2
	// blocks.
	time.Sleep(time.Duration(consensus.NumSteps) * 2 * consensus.StepDuration)

	// Query to verify that the request was accepted by the network.
	err = c.router.SendMessage(network.Message{
		Dest: c.metadata.Siblings[0].Address,
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

	c.genericWallets[id] = gw

	return
}

// Iterates through the client and saves all of the wallets to disk.
func (c *Client) SaveAllWallets() (err error) {
	var filename string
	for id, keypair := range c.genericWallets {
		filename = fmt.Sprintf("%x.id", id)
		err = SaveWallet(id, keypair, filename)
		if err != nil {
			return
		}
	}
	// Save other types of wallets as they are implemented.
	return
}

func SaveWallet(id state.WalletID, keypair GenericWallet, destFile string) (err error) {
	f, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(siaencoding.EncUint64(uint64(id)))
	if err != nil {
		return
	}
	_, err = f.Write(keypair.PublicKey[:])
	if err != nil {
		return
	}
	_, err = f.Write(keypair.SecretKey[:])
	if err != nil {
		return
	}
	return
}

func LoadWallet(fileName string) (id state.WalletID, keypair GenericWallet, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	idSlice := make([]byte, 8)
	pubSlice := make([]byte, siacrypto.PublicKeySize)
	secSlice := make([]byte, siacrypto.SecretKeySize)
	_, err = f.Read(idSlice)
	if err != nil {
		return
	}
	_, err = f.Read(pubSlice)
	if err != nil {
		return
	}
	_, err = f.Read(secSlice)
	if err != nil {
		return
	}
	id = state.WalletID(siaencoding.DecUint64(idSlice))
	if copy(keypair.PublicKey[:], pubSlice) != siacrypto.PublicKeySize {
		err = errors.New("bad public key length")
		return
	}
	if copy(keypair.SecretKey[:], secSlice) != siacrypto.SecretKeySize {
		err = errors.New("bad secret key length")
		return
	}

	return
}
