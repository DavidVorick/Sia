package main

import (
	"math/rand"

	//"github.com/NebulousLabs/Sia/network"

	"github.com/nsf/termbox-go"
)

// Draw the wallets section in the priary screen.
func drawWallets(startColumn int) {
	// Fetch a list of wallets from the server.

	// Fill remaining space with random colors.
	for y := 0; y < context.Height; y++ {
		for x := startColumn; x < context.Width; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+1)
		}
	}
}