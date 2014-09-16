package main

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

type SettingsView struct {
	Rectangle
	Parent   View
	hasFocus bool
}

func (sv *SettingsView) SetDims(r Rectangle) {
	sv.Rectangle = r
}

func (sv *SettingsView) GiveFocus() {
	sv.hasFocus = true
}

// Draw the wallets section in the priary screen.
func (sv *SettingsView) Draw() {
	// Fill remaining space with random colors.
	for y := sv.MinY; y < sv.MaxY; y++ {
		for x := sv.MinX; x < sv.MaxX; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+5)
		}
	}
}

func (sv *SettingsView) HandleKey(key termbox.Key) {

}
