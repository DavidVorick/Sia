package main

import (
	"github.com/nsf/termbox-go"
)

const (
	DividerWidth = 1
	DividerColor = termbox.ColorBlue
)

// An MVC comprises methods conducive to the MVC design pattern.
// The Model methods modify the internal state of the object.
// The View methods dictate how the object is presented on-screen.
// The Controller methods handle user input.
//
// This client is designed as a hierarchy of such MVCs. Parents commonly
// forward user input to a child's Controller, and children commonly return
// user focus to their parent.
type MVC interface {
	// Model
	Focus()
	// View
	SetDims(Rectangle)
	Draw()
	// Controller
	HandleKey(termbox.Key)
	HandleRune(rune)
}

// The DefaultMVC contains fields common to most MVCs.
// It also implements a very basic MVC interface, to cut down on
// boilerplate code.
type DefaultMVC struct {
	Rectangle
	Parent   MVC
	hasFocus bool
}

// Bare-bones implementation of the MVC interface
func (d *DefaultMVC) SetDims(r Rectangle)   { d.Rectangle = r }
func (d *DefaultMVC) Focus()                { d.hasFocus = true }
func (d *DefaultMVC) Draw()                 {}
func (d *DefaultMVC) HandleKey(termbox.Key) {}
func (d *DefaultMVC) HandleRune(rune)       {}

// GiveFocus is a helper function that removes focus from the current MVC and
// focuses its argument. Since only one MVC should have focus at any given
// time, it checks that the current MVC has focus before transferring it.
func (d *DefaultMVC) GiveFocus(target MVC) {
	if !d.hasFocus {
		panic("focus is not yours to give!")
	}
	d.hasFocus = false
	target.Focus()
}

// A MenuMVC is a navigable menu and viewing window, vertically separated.
// Because the window is a MVC and MenuMVC implements the MVC interface,
// MenuMVCs can be nested.
type MenuMVC struct {
	DefaultMVC
	Title     string
	MenuWidth int
	Items     []string
	Windows   []MVC
	sel       int
}

func (m *MenuMVC) SetDims(r Rectangle) {
	m.Rectangle = r
	r.MinX += m.MenuWidth + DividerWidth
	for i := range m.Windows {
		m.Windows[i].SetDims(r)
	}
}

func (m *MenuMVC) Focus() {
	m.hasFocus = true
	// return focus to parent if we have nothing to highlight
	if len(m.Items) == 0 {
		m.GiveFocus(m.Parent)
	}
}

func (m *MenuMVC) Draw() {
	// draw title and divider
	drawColorString(m.MinX+1, m.MinY+1, m.Title, HomeHeaderColor, termbox.ColorDefault)
	for y := m.MinY; y < m.MaxY; y++ {
		termbox.SetCell(m.MinX+m.MenuWidth, y, '│', DividerColor, termbox.ColorDefault)
	}

	if len(m.Items) == 0 {
		drawString(m.MinX+1, m.MinY+3, "<empty>")
		return
	}

	// draw menu items
	for i, s := range m.Items {
		drawColorString(m.MinX+1, m.MinY+2*i+3, s, HomeInactiveColor, termbox.ColorDefault)
	}
	// highlight selected item
	if m.hasFocus {
		drawLine(m.MinX, m.MinY+2*m.sel+3, m.MenuWidth, HomeActiveColor)
		drawColorString(m.MinX+1, m.MinY+2*m.sel+3, m.Items[m.sel], termbox.ColorWhite, HomeActiveColor)
	} else {
		drawLine(m.MinX, m.MinY+2*m.sel+3, m.MenuWidth, HomeInactiveColor)
		drawColorString(m.MinX+1, m.MinY+2*m.sel+3, m.Items[m.sel], termbox.ColorWhite, HomeInactiveColor)
	}

	// draw current window
	m.Windows[m.sel].Draw()
}

// If the current focus is on the window (instead of the menu), the input is
// forwarded to the current subview. This is a common pattern in MVCs that
// contain subviews.
func (m *MenuMVC) HandleKey(key termbox.Key) {
	if !m.hasFocus {
		m.Windows[m.sel].HandleKey(key)
		return
	}

	switch key {
	case termbox.KeyArrowUp:
		if m.sel > 0 {
			m.sel--
		}
	case termbox.KeyArrowDown:
		if m.sel+1 < len(m.Items) {
			m.sel++
		}
	case termbox.KeyArrowLeft:
		if m.Parent != nil {
			m.GiveFocus(m.Parent)
		}
	case termbox.KeyArrowRight:
		if len(m.Windows) > m.sel {
			m.GiveFocus(m.Windows[m.sel])
		}
	default:
		drawError("Invalid key")
	}
}

func (m *MenuMVC) HandleRune(r rune) {
	if !m.hasFocus {
		m.Windows[m.sel].HandleRune(r)
		return
	}
}
