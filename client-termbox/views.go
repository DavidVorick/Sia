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
func (dv *DefaultView) SetDims(r Rectangle) { dv.Rectangle = r }
func (dv *DefaultView) Focus()              { dv.hasFocus = true }
func (dv *DefaultView) Draw()               {}
func (dv *DefaultView) HandleKey(key termbox.Key) {
	if key == termbox.KeyArrowLeft && dv.Parent != nil {
		dv.GiveFocus(dv.Parent)
	}
}
func (dv *DefaultView) HandleRune(rune) {}

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

// Draw implements the View.Draw method, drawing the MenuView inside the
// given rectangle.
func (mw *MenuView) Draw() {
	// draw menu
	drawString(mw.MinX+1, mw.MinY+1, mw.Title, HomeHeaderColor, termbox.ColorDefault)
	for i, s := range mw.Items {
		drawString(mw.MinX+1, mw.MinY+2*i+3, s, HomeInactiveColor, termbox.ColorDefault)
	}
	// highlight selected item
	if mw.hasFocus {
		drawLine(mw.MinX, mw.MinY+2*mw.sel+3, mw.MenuWidth, HomeActiveColor)
		drawString(mw.MinX+1, mw.MinY+2*mw.sel+3, mw.Items[mw.sel], termbox.ColorWhite, HomeActiveColor)
	} else {
		drawString(mw.MinX+1, mw.MinY+2*mw.sel+3, mw.Items[mw.sel], termbox.ColorWhite, termbox.ColorDefault)
	}

	// draw divider
	for y := mw.MinY; y < mw.MaxY; y++ {
		termbox.SetCell(mw.MinX+mw.MenuWidth, y, '│', DividerColor, termbox.ColorDefault)
	}

	// draw window
	mw.Windows[mw.sel].Draw()
}

// HandleKey implements the View.HandleKey method. If the current focus is on
// the window (instead of the menu), the input is forwarded to the window View.
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
		mw.GiveFocus(mw.Windows[mw.sel])
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
