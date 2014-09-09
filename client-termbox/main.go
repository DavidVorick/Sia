package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	HomeBoxWidth = 15
	Border       = 1

	HomeHeaderColor   = termbox.ColorRed
	HomeActiveColor   = termbox.ColorBlue
	HomeInactiveColor = termbox.ColorGreen

	DividerColor = termbox.ColorBlue
)

type Context struct {
	WalletsActive      bool
	ParticipantsActive bool
	SettingsActive     bool
}

var context Context

// Draw uses the context field to determine what functions to call when drawing
// the image. Siabox uses a box-style of programming, each function receives a
// box in which it can draw things, and is given an offset so that it knows
// where that box is.
func draw() {
	// Get size of whole window.
	w, h := termbox.Size()

	// Determine how to draw the home field.
	var homeSeparator int
	if w <= HomeBoxWidth {
		// If there isn't enough room for the home box, then just draw
		// the home box as red. This will be context dependent, the
		// program will try to have enough room for whatever box the
		// context says is active.
	} else {
		homeSeparator = HomeBoxWidth
	}

	// Draw the home box.
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < h; y++ {
		termbox.SetCell(homeSeparator, y, ' ', termbox.ColorDefault, DividerColor)
	}

	// Write the home fields.
	for x, c := range "Sia Alpha v3" {
		termbox.SetCell(x+Border, Border, c, HomeHeaderColor, termbox.ColorDefault)
	}
	for x, c := range "Wallets" {
		if context.WalletsActive {
			termbox.SetCell(x+Border, 2*Border+1, c, HomeActiveColor, termbox.ColorDefault)
		} else {
			termbox.SetCell(x+Border, 2*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}
	for x, c := range "Participants" {
		if context.ParticipantsActive {
			termbox.SetCell(x+Border, 4*Border+1, c, HomeActiveColor, termbox.ColorDefault)
		} else {
			termbox.SetCell(x+Border, 4*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}
	for x, c := range "Settings" {
		if context.SettingsActive {
			termbox.SetCell(x+Border, 6*Border+1, c, HomeActiveColor, termbox.ColorDefault)
		} else {
			termbox.SetCell(x+Border, 6*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}

	// Fill remaining space with random colors.
	for y := 0; y < h; y++ {
		for x := homeSeparator + 1; x < w; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute(rand.Int()%8)+1)
		}
	}
	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer termbox.Close()

	event_queue := make(chan termbox.Event)
	go func() {
		for {
			event_queue <- termbox.PollEvent()
		}
	}()

	draw()

	for {
		select {
		case ev := <-event_queue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				return
			}
		default:
			draw()
			time.Sleep(100 * time.Millisecond)
		}
	}
}
