package main

import (
	"github.com/nsf/termbox-go"
)

// Home is a special box without offsets, because it always starts at 0,0, and
// ends at width, height
func drawHome() {
	// Determine how to draw the home field.
	var homeSeparator int
	if width <= HomeBoxWidth {
		// If there isn't enough room for the home box, then just draw
		// the home box as red. This will be context dependent, the
		// program will try to have enough room for whatever box the
		// context says is active.
	} else {
		homeSeparator = HomeBoxWidth
	}

	// Draw the home box.
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < height; y++ {
		termbox.SetCell(homeSeparator, y, ' ', termbox.ColorDefault, DividerColor)
	}

	// Write the home fields.
	for x, c := range "Sia Alpha v3" {
		termbox.SetCell(x+Border, Border, c, HomeHeaderColor, termbox.ColorDefault)
	}
	for x, c := range "Wallets" {
		if context.WalletsActive {
			termbox.SetCell(x+Border, 2*Border+1, c, HomeActiveColor, termbox.ColorDefault)
			drawWallets(homeSeparator + 1 + Border)
		} else {
			termbox.SetCell(x+Border, 2*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}
	for x, c := range "Participants" {
		if context.ParticipantsActive {
			termbox.SetCell(x+Border, 4*Border+1, c, HomeActiveColor, termbox.ColorDefault)
			drawParticipants(homeSeparator + 1 + Border)
		} else {
			termbox.SetCell(x+Border, 4*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}
	for x, c := range "Settings" {
		if context.SettingsActive {
			termbox.SetCell(x+Border, 6*Border+1, c, HomeActiveColor, termbox.ColorDefault)
			drawSettings(homeSeparator + 1 + Border)
		} else {
			termbox.SetCell(x+Border, 6*Border+1, c, HomeInactiveColor, termbox.ColorDefault)
		}
	}

}

func homeEvent(event termbox.Event) {
	if event.Type == termbox.EventKey {
		if event.Key == termbox.KeyArrowUp {
			if context.ParticipantsActive {
				context.ParticipantsActive = false
				context.WalletsActive = true
			} else if context.SettingsActive {
				context.SettingsActive = false
				context.ParticipantsActive = true
			}
		} else if event.Key == termbox.KeyArrowDown {
			if context.WalletsActive {
				context.WalletsActive = false
				context.ParticipantsActive = true
			} else if context.ParticipantsActive {
				context.ParticipantsActive = false
				context.SettingsActive = true
			}
		}
	}
}
