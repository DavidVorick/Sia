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
	mw.GiveFocus()
	w, h := termbox.Size()

	for {
		// update view
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		mw.Draw(Rectangle{0, 0, w, h})
		termbox.Flush()

		// handle next event
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			if event.Key == termbox.KeyEsc {
				return
			}
			mw.HandleKey(event.Key)

		case termbox.EventResize:
			w, h = termbox.Size()

		case termbox.EventMouse:
			// mouse events not yet supported

		case termbox.EventError:
			fmt.Println("Input error:", event.Err)
			return
		}
	}
}
