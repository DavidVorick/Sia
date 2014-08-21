package client

import (
	"errors"
	"time"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/delta"
	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

/*
// send a user-specified script input
func (c *Client) SendCustomInput(id state.WalletID, input []byte) (err error) {
	return c.router.SendMessage(network.Message{
		Dest: c.connectAddress,
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: id,
			Input:    input,
		},
		Resp: nil,
	})
}
*/

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

	// Send the requesting script input out to the network.
	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: delta.FountainWalletID,
			Input:    delta.CreateFountainWalletInput(id, delta.DefaultScript(pk)),
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

/*
// send coins from one wallet to another
func (c *Client) SubmitTransaction(src, dst state.WalletID, amount uint64) (err error) {
	_, exists := c.genericWallets[src]
	if !exists {
		err = fmt.Errorf("Could not access source wallet")
		return
	}

	input := delta.TransactionInput(uint64(dst), 0, amount)
	input, err = delta.SignInput(c.genericWallets[src].SK, input)
	if err != nil {
		return
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: src,
			Input:    input,
		},
		Resp: nil,
	})
	return
}
*/

/*
// resize sector associated with wallet
func (c *Client) ResizeSector(w state.WalletID, atoms uint16, k byte) (err error) {
	_, exists := c.genericWallets[src]
	if !exists {
		err = fmt.Errorf("Could not access wallet")
		return
	}

	input := delta.ResizeSectorEraseInput(atoms, k)
	input, err = delta.SignInput(c.genericWallets[w].SK, input)
	if err != nil {
		return
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: w,
			Input:    input,
		},
		Resp: nil,
	})
	return
}
*/
