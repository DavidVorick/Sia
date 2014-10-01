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
	Focus()
	Draw()
	HandleKey(termbox.Key)
	HandleRune(rune)
}

// The DefaultView contains fields common to most Views.
// It also implements a very basic View interface, to cut down on
// boilerplate code.
type DefaultView struct {
	Rectangle
	Parent   View
	hasFocus bool
}

// Bare-bones implementation of the View interface
func (dv *DefaultView) SetDims(r Rectangle)   { dv.Rectangle = r }
func (dv *DefaultView) Focus()                { dv.hasFocus = true }
func (dv *DefaultView) Draw()                 {}
func (dv *DefaultView) HandleKey(termbox.Key) {}
func (dv *DefaultView) HandleRune(rune)       {}

// GiveFocus is a helper function that removes focus from the current View and
// focuses its argument. Since only one View should have focus at any given
// time, it checks that the current View has focus before transferring it.
func (dv *DefaultView) GiveFocus(v View) {
	if !dv.hasFocus {
		panic("focus is not yours to give!")
	}
	dv.hasFocus = false
	v.Focus()
}

// A MenuView is a navigable menu and viewing window, vertically separated.
// Because the window is a View and MenuView implements the View interface,
// MenuViews can be nested.
type MenuView struct {
	DefaultView
	Title     string
	MenuWidth int
	Items     []string
	Windows   []View
	sel       int
}

func (mw *MenuView) SetDims(r Rectangle) {
	mw.Rectangle = r
	r.MinX += mw.MenuWidth + DividerWidth
	for i := range mw.Windows {
		mw.Windows[i].SetDims(r)
	}
}

func (mw *MenuView) Focus() {
	mw.hasFocus = true
	// return focus to parent if we have nothing to highlight
	if len(mw.Items) == 0 {
		mw.GiveFocus(mw.Parent)
	}
}

func (mw *MenuView) Draw() {
	// draw title and divider
	drawColorString(mw.MinX+1, mw.MinY+1, mw.Title, HomeHeaderColor, termbox.ColorDefault)
	for y := mw.MinY; y < mw.MaxY; y++ {
		termbox.SetCell(mw.MinX+mw.MenuWidth, y, '│', DividerColor, termbox.ColorDefault)
	}

	if len(mw.Items) == 0 {
		drawString(mw.MinX+1, mw.MinY+3, "<empty>")
		return
	}

	// draw menu items
	for i, s := range mw.Items {
		drawColorString(mw.MinX+1, mw.MinY+2*i+3, s, HomeInactiveColor, termbox.ColorDefault)
	}
	// highlight selected item
	if mw.hasFocus {
		drawLine(mw.MinX, mw.MinY+2*mw.sel+3, mw.MenuWidth, HomeActiveColor)
		drawColorString(mw.MinX+1, mw.MinY+2*mw.sel+3, mw.Items[mw.sel], termbox.ColorWhite, HomeActiveColor)
	} else {
		drawLine(mw.MinX, mw.MinY+2*mw.sel+3, mw.MenuWidth, HomeInactiveColor)
		drawColorString(mw.MinX+1, mw.MinY+2*mw.sel+3, mw.Items[mw.sel], termbox.ColorWhite, HomeInactiveColor)
	}

	// draw current window
	mw.Windows[mw.sel].Draw()
}

// If the current focus is on the window (instead of the menu), the input is
// forwarded to the current subview. This is a common pattern in Views that
// contain subviews.
func (mw *MenuView) HandleKey(key termbox.Key) {
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
	case termbox.KeyArrowLeft:
		if mw.Parent != nil {
			mw.GiveFocus(mw.Parent)
		}
	case termbox.KeyArrowRight:
		if len(mw.Windows) > mw.sel {
			mw.GiveFocus(mw.Windows[mw.sel])
		}
	default:
		//drawError("Invalid key")
	}
}

func (mw *MenuView) HandleRune(r rune) {
	if !mw.hasFocus {
		mw.Windows[mw.sel].HandleRune(r)
		return
	}
}
