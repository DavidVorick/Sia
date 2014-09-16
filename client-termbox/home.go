package main

import (
	"fmt"
	"time"

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
		Menu: struct {
			Width int
			Title string
			Items []string
		}{
			HomeMenuWidth,
			"Sia Alpha v3",
			[]string{"Wallets", "Participants", "Settings"},
		},
		sel:      0,
		winFocus: false,
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
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()

	// create main window
	mw := newHomeView()
	w, h := termbox.Size()

	// redraw screen at 10 Hz forever
	go func() {
		for {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			mw.Draw(Rectangle{0, 0, w, h})
			termbox.Flush()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// handle input until exit
	for {
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
