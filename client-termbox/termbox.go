package main

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	HomeBoxWidth = 15
	Border       = 1

	HomeHeaderColor   = termbox.ColorRed
	HomeActiveColor   = termbox.ColorYellow
	HomeInactiveColor = termbox.ColorGreen

	DividerColor = termbox.ColorBlue
)

type Context struct {
	Width  int
	Height int

	Focus string

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
	context.Width, context.Height = termbox.Size()
	drawHome()
	termbox.Flush()
}

func termboxRun() {
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

	context.Focus = "home"
	context.WalletsActive = true

	draw()

	for {
		select {
		case event := <-event_queue:
			// Check for the quit signal.
			if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
				return
			}

			switch context.Focus {
			case "home":
				homeEvent(event)
			default:
				panic("focus not home!") // Panic because focus should never be off of home yet!
			}
		default:
			draw()
			time.Sleep(25 * time.Millisecond)
		}
	}
}
