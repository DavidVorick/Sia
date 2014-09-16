package main

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

type SettingsView struct {
	Parent   View
	hasFocus bool
}

// Draw the wallets section in the priary screen.
func (sv *SettingsView) Draw(r Rectangle) {
	// Fill remaining space with random colors.
	for y := r.MinY; y < r.MaxY; y++ {
		for x := r.MinX; x < r.MaxX; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+5)
		}
	}
}

func (sv *SettingsView) HandleKey(key termbox.Key) {

}

func (sv *SettingsView) GiveFocus() {
	sv.hasFocus = true
}
