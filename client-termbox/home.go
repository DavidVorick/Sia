package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	HomeMenuWidth     = 15
	HomeHeaderColor   = termbox.ColorRed
	HomeActiveColor   = termbox.ColorYellow
	HomeInactiveColor = termbox.ColorGreen
)

func newHomeView() View {
	// create MenuWindow
	mw := &MenuWindow{
		Title:     "Sia Alpha v3",
		MenuWidth: HomeMenuWidth,
		Items:     []string{"Wallets", "Participants", "Settings"},
		sel:       0,
		hasFocus:  false,
	}

	// add subviews
	mw.Windows = []View{
		&WalletsView{Parent: mw},
		&ParticipantsView{Parent: mw},
		&SettingsView{Parent: mw},
	}

	return mw
}

func termboxRun() {
	if err := termbox.Init(); err != nil {
		fmt.Println(err)
		return
	}
	defer termbox.Close()

	// create main window
	mw := newHomeView()
	w, h := termbox.Size()
	mw.SetDims(Rectangle{0, 0, w, h})
	mw.GiveFocus()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	mw.Draw()
	termbox.Flush()

	for {
		// handle next event
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			if event.Key == termbox.KeyEsc {
				return
			}
			mw.HandleKey(event.Key)

		case termbox.EventResize:
			w, h = event.Width, event.Height
			mw.SetDims(Rectangle{0, 0, w, h})
			mw.Draw()

		case termbox.EventMouse:
			// mouse events not yet supported

		case termbox.EventError:
			drawError("Input error:", event.Err)
			termbox.Flush()
			return
		}

		// update view
		termbox.Flush()
	}
}
