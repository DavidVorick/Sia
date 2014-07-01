package client

import (
	"fmt"
	"network"
	"participant"
	"quorum"
	"quorum/script"
	"siacrypto"
)

// send a user-specified script input
func (c *Client) SendCustomInput(id quorum.WalletID, input []byte) (err error) {
	return c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: id,
			Input:    input,
		},
		Resp: nil,
	})
}

// request a new wallet from the bootstrap
func (c *Client) RequestWallet(id quorum.WalletID, s []byte) (err error) {
	// Create a generic wallet with a keypair for the request
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}
	c.genericWallets[id] = new(siacrypto.Keypair)
	c.genericWallets[id].PK = pk
	c.genericWallets[id].SK = sk

	// use default script if none provided
	if s == nil {
		s = script.CreateWalletInput(uint64(id), script.DefaultScript(c.genericWallets[id].PK))
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: participant.BootstrapID,
			Input:    s,
		},
		Resp: nil,
	})
	return
}

// send coins from one wallet to another
func (c *Client) SubmitTransaction(src, dst quorum.WalletID, amount uint64) (err error) {
	if c.genericWallets[src] == nil {
		err = fmt.Errorf("Could not access source wallet")
		return
	}

	input := script.TransactionInput(uint64(dst), 0, amount)
	input, err = script.SignInput(c.genericWallets[src].SK, input)
	if err != nil {
		return
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: src,
			Input:    input,
		},
		Resp: nil,
	})
	return
}

// resize sector associated with wallet
func (c *Client) ResizeSector(w quorum.WalletID, atoms uint16, k byte) (err error) {
	if c.genericWallets[w] == nil {
		err = fmt.Errorf("Could not access wallet")
		return
	}

	input := script.ResizeSectorEraseInput(atoms, k)
	input, err = script.SignInput(c.genericWallets[w].SK, input)
	if err != nil {
		return
	}

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: w,
			Input:    input,
		},
		Resp: nil,
	})
	return
}
