package main

import (
	"github.com/nsf/termbox-go"
)

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
