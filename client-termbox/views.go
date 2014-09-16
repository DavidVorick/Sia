package main

import (
	"github.com/nsf/termbox-go"
)

const (
	DividerWidth = 1
	DividerColor = termbox.ColorBlue
)

// A View is an area on screen capable of drawing itself and handling input.
type View interface {
	Draw(Rectangle)
	HandleKey(termbox.Key)
	//GiveFocus(View)
}

// A MenuWindow is a navigable menu and viewing window, vertically separated.
// Because the window is a View and MenuWindow implements the View interface,
// MenuWindows can be nested.
type MenuWindow struct {
	Title     string
	MenuWidth int
	Items     []string
	Windows   []View
	sel       int
	hasFocus  bool
}

// Draw implements the View.Draw method, drawing the MenuWindow inside the
// given rectangle.
func (mw *MenuWindow) Draw(r Rectangle) {
	// draw menu
	drawString(r.MinX+1, r.MinY+1, mw.Title, HomeHeaderColor, termbox.ColorDefault)
	for i, s := range mw.Items {
		if i == mw.sel {
			drawString(r.MinX+1, r.MinY+2*i+3, s, HomeActiveColor, termbox.ColorDefault)
		} else {
			drawString(r.MinX+1, r.MinY+2*i+3, s, HomeInactiveColor, termbox.ColorDefault)
		}
	}

	// draw divider
	for y := r.MinY; y < r.MaxY; y++ {
		termbox.SetCell(mw.MenuWidth, y, '│', DividerColor, termbox.ColorDefault)
	}

	// draw window
	r.MinX += mw.MenuWidth + DividerWidth
	mw.Windows[mw.sel].Draw(r)
}

// HandleKey implements the view.HandleKey method. If the current focus is on
// the window (instead of the menu), the input is forwarded to the window View.
func (mw *MenuWindow) HandleKey(key termbox.Key) {
	if !mw.hasFocus {
		mw.Windows[mw.sel].HandleKey(key)
		return
	}

	switch key {
	case termbox.KeyArrowUp:
		if mw.sel > 0 {
			mw.sel--
		}
	case termbox.KeyArrowDown:
		if mw.sel+1 < len(mw.Items) {
			mw.sel++
		}
	default:
		//drawError("Invalid key")
	}
}
