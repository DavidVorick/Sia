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
		Items: []string{
			"Wallets",
			"Participants",
			"Settings",
		},
	}

	// add subviews
	mw.Windows = []View{
		newWalletMenuView(mw),
		&ParticipantsView{DefaultView{Parent: mw}},
		newSettingsView(mw),
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
	mw.Focus()

	for {
		// update view
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		mw.Draw()
		termbox.Flush()

		// handle next event
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			switch {
			case event.Ch != 0:
				mw.HandleChar(event.Ch)
			case event.Key == termbox.KeyEsc:
				return
			default:
				mw.HandleKey(event.Key)
			}

		case termbox.EventResize:
			w, h = event.Width, event.Height
			mw.SetDims(Rectangle{0, 0, w, h})

		case termbox.EventMouse:
			// mouse events not yet supported

		case termbox.EventError:
			//drawError("Input error:", event.Err)
			termbox.Flush()
			return
		}
	}
}
