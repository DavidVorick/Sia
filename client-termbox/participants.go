package main

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

type ParticipantsView struct {
	Parent   View
	hasFocus bool
}

// Draw the wallets section in the priary screen.
func (pv *ParticipantsView) Draw(r Rectangle) {
	// Fill remaining space with random colors.
	for y := r.MinY; y < r.MaxY; y++ {
		for x := r.MinX; x < r.MaxX; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+3)
		}
	}
}

func (pv *ParticipantsView) HandleKey(key termbox.Key) {

}

func (pv *ParticipantsView) GiveFocus() {
	pv.hasFocus = true
}
