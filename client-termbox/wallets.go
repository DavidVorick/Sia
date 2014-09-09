package main

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

// Draw the wallets section in the priary screen.
func drawWallets(startColumn int) {
	// Fill remaining space with random colors.
	for y := 0; y < height; y++ {
		for x := startColumn; x < width; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+1)
		}
	}
}
