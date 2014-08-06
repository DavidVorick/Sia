package client

import (
	"delta"
	"network"
	"state"
)

// send a user-specified script input
func (c *Client) SendCustomInput(id state.WalletID, input []byte) (err error) {
	return c.router.SendMessage(network.Message{
		Dest: c.bootstrap,
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: id,
			Input:    input,
		},
		Resp: nil,
	})
}

/*
// Submit a wallet request to the bootstrap wallet.
func (c *Client) RequestWallet(id state.WalletID, s []byte) (err error) {
	// Create a generic wallet with a keypair for the request.
	pk, sk, err := siacrypto.CreateKeyPair()
	if err != nil {
		return
	}

	// Fill out a keypair object and insert it into the generic wallet map.
	var kp Keypair
	kp.PK = pk
	kp.SK = sk
	c.genericWallets[id] = kp

	c.Broadcast(network.Message{
		Proc: "Participant.AddScriptInput",
		Args: delta.ScriptInput{
			WalletID: delta.BootstrapWalletID,
			Input:    s,
		},
		Resp: nil,
	})
	return
}
*/

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
