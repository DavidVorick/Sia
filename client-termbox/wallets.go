package main

import (
	"math/rand"

	"github.com/NebulousLabs/Sia/network"
	"github.com/NebulousLabs/Sia/state"

	"github.com/nsf/termbox-go"
)

// Draw the wallets section in the priary screen.
func drawWallets(startColumn int) {
	// Fetch a list of wallets from the server.
	var ids *[]state.WalletID
	err := networkState.Router.SendMessage(network.Message{
		Dest: networkState.ServerAddress,
		Proc: "Server.WalletIDs",
		Args: struct{}{},
		Resp: &ids,
	})
	if err != nil {
		// Draw an error message.
		return
	}

	// Fill remaining space with random colors.
	for y := 0; y < context.Height; y++ {
		for x := startColumn; x < context.Width; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+1)
		}
	}
}
