package main

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

type ParticipantsView struct {
	DefaultView
}

func (pv *ParticipantsView) SetDims(r Rectangle) {
	pv.Rectangle = r
}

func (pv *ParticipantsView) GiveFocus() {
	pv.hasFocus = true
}

// Draw the wallets section in the priary screen.
func (pv *ParticipantsView) Draw() {
	// Fill remaining space with random colors.
	for y := pv.MinY; y < pv.MaxY; y++ {
		for x := pv.MinX; x < pv.MaxX; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%2)+3)
		}
	}
}

func (pv *ParticipantsView) HandleKey(key termbox.Key) {

}
