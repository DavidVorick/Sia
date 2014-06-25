package client

import (
	"fmt"
	"network"
	"participant"
	"quorum"
	"quorum/script"
	"siacrypto"
)

// request a new wallet from the bootstrap
func (c *Client) RequestWallet(id quorum.WalletID) (err error) {
	// Create a generic wallet with a keypair for the request
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}
	c.genericWallets[id] = new(siacrypto.Keypair)
	c.genericWallets[id].PK = pk
	c.genericWallets[id].SK = sk

	return c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: participant.BootstrapID,
			Input:    script.CreateWalletInput(uint64(id), script.DefaultScript(c.genericWallets[id].PK)),
		},
		Resp: nil,
	})
}

// send coins from one wallet to another
func (c *Client) SubmitTransaction(src, dst quorum.WalletID, amount uint64) (err error) {
	if c.genericWallets[src] == nil {
		err = fmt.Errorf("Could not access source wallet")
		return
	}

	sm, err := script.RandomSignedMessage(c.genericWallets[src].SK)
	if err != nil {
		return
	}

	return c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: src,
			Input:    script.TransactionInput(sm, uint64(dst), 0, amount),
		},
		Resp: nil,
	})
}

// resize sector associated with wallet
func (c *Client) ResizeSector(w quorum.WalletID, atoms uint16, m byte) (err error) {
	if c.genericWallets[w] == nil {
		err = fmt.Errorf("Could not access wallet")
		return
	}

	sm, err := script.RandomSignedMessage(c.genericWallets[w].SK)
	if err != nil {
		return
	}
	return c.router.SendMessage(&network.Message{
		Dest: participant.BootstrapAddress,
		Proc: "Participant.AddScriptInput",
		Args: script.ScriptInput{
			WalletID: w,
			Input:    script.ResizeSectorEraseInput(sm, atoms, m),
		},
		Resp: nil,
	})
}
