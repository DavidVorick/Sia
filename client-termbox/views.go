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
	SetDims(Rectangle)
	GiveFocus()
	Draw()
	HandleKey(termbox.Key)
}

// A MenuWindow is a navigable menu and viewing window, vertically separated.
// Because the window is a View and MenuWindow implements the View interface,
// MenuWindows can be nested.
type MenuWindow struct {
	Rectangle
	Parent    View
	Title     string
	MenuWidth int
	Items     []string
	Windows   []View
	sel       int
	hasFocus  bool
}

func (mw *MenuWindow) SetDims(r Rectangle) {
	mw.Rectangle = r
	r.MinX += mw.MenuWidth + DividerWidth
	for i := range mw.Windows {
		mw.Windows[i].SetDims(r)
	}
}

// GiveFocus implements the View.GiveFocus method.
func (mw *MenuWindow) GiveFocus() {
	mw.hasFocus = true
}

// Draw implements the View.Draw method, drawing the MenuWindow inside the
// given rectangle.
func (mw *MenuWindow) Draw() {
	// draw menu
	clearRectangle(mw.Rectangle)
	drawString(mw.MinX+1, mw.MinY+1, mw.Title, HomeHeaderColor, termbox.ColorDefault)
	for i, s := range mw.Items {
		drawString(mw.MinX+1, mw.MinY+2*i+3, s, HomeInactiveColor, termbox.ColorDefault)
	}
	// highlight selected item
	drawString(mw.MinX+1, mw.MinY+2*mw.sel+3, mw.Items[mw.sel], HomeActiveColor, termbox.ColorDefault)

	// draw divider
	for y := mw.MinY; y < mw.MaxY; y++ {
		termbox.SetCell(mw.MenuWidth, y, '│', DividerColor, termbox.ColorDefault)
	}

	// draw window
	mw.Windows[mw.sel].Draw()
}

// HandleKey implements the View.HandleKey method. If the current focus is on
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
		mw.Draw()
	case termbox.KeyArrowDown:
		if mw.sel+1 < len(mw.Items) {
			mw.sel++
		}
		mw.Draw()
	case termbox.KeyArrowLeft:
		if mw.Parent != nil {
			mw.hasFocus = false
			mw.Parent.GiveFocus()
			mw.Parent.Draw()
		}
	case termbox.KeyArrowRight:
		mw.hasFocus = false
		mw.Windows[mw.sel].GiveFocus()
		mw.Draw()
	default:
		drawError("Invalid key")
	}
}
