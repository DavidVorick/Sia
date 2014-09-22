package main

import (
	"github.com/nsf/termbox-go"
)

const (
	SettingColor      = termbox.ColorBlue
	SettingFocusColor = termbox.ColorRed
)

type Field struct {
	DefaultView
	color termbox.Attribute
	text  string
	pos   int
}

func (f *Field) Focus() {
	f.hasFocus = true
	f.pos = len(f.text)
	termbox.SetCursor(f.MinX+f.pos, f.MinY)
}

func (f *Field) Draw() {
	drawRectangle(f.Rectangle, f.color)
	drawString(f.MinX, f.MinY, f.text, termbox.ColorWhite, f.color)
}

func (f *Field) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyEnter:
		termbox.HideCursor()
		f.GiveFocus(f.Parent)
	case termbox.KeyArrowLeft:
		if f.pos > 0 {
			f.pos--
		}
		f.updateCursor()
	case termbox.KeyArrowRight:
		if f.pos < len(f.text) {
			f.pos++
		}
		f.updateCursor()
	case termbox.KeySpace:
		f.HandleRune(' ')
	case termbox.KeyTab:
		f.HandleRune('\t')
	case termbox.KeyDelete:
		if f.pos < len(f.text) {
			f.deleteForward()
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if f.pos > 0 {
			f.deleteBackward()
			f.pos--
			f.updateCursor()
		}
	}
}

func (f *Field) HandleRune(r rune) {
	if len(f.text) >= f.MaxX-f.MinX-1 {
		return
	}
	f.text = f.text[:f.pos] + string(r) + f.text[f.pos:]
	f.pos++
	f.updateCursor()
}

func (f *Field) updateCursor() {
	termbox.SetCursor(f.MinX+f.pos, f.MinY)
}

func (f *Field) deleteForward() {
	f.text = f.text[:f.pos] + f.text[f.pos+1:]
}

func (f *Field) deleteBackward() {
	f.text = f.text[:f.pos-1] + f.text[f.pos:]
}

func (f *Field) SetColor(color termbox.Attribute) {
	f.color = color
}

func newField(parent View, color termbox.Attribute, text string) View {
	f := &Field{
		color: color,
		text:  text,
	}
	f.Parent = parent
	return f
}

type Setting struct {
	Rectangle
	Field
	name   string
	width  int
	offset int
}

func (s *Setting) SetDims(r Rectangle) {
	s.Rectangle = r
	r.MinX += len(s.name) + 1
	s.Field.SetDims(r)
}

func (s *Setting) Draw() {
	drawString(s.MinX, s.MinY, s.name, termbox.ColorWhite, termbox.ColorDefault)
	s.Field.Draw()
}

type InputView struct {
	DefaultView
	settings []*Setting
	sel      int
}

func (sv *InputView) SetDims(r Rectangle) {
	sv.Rectangle = r
	for _, s := range sv.settings {
		s.SetDims(Rectangle{
			MinX: r.MinX + 1,
			MinY: r.MinY + s.offset,
			MaxX: r.MinX + len(s.name) + s.width + 2,
			MaxY: r.MinY + s.offset + 1,
		})
	}
}

func (sv *InputView) Draw() {
	for i, s := range sv.settings {
		if i == sv.sel && sv.hasFocus {
			s.SetColor(SettingFocusColor)
		} else {
			s.SetColor(SettingColor)
		}
		s.Draw()
	}
}

func (sv *InputView) HandleKey(key termbox.Key) {
	if !sv.hasFocus {
		sv.settings[sv.sel].HandleKey(key)
		return
	}
	switch key {
	case termbox.KeyArrowLeft:
		sv.GiveFocus(sv.Parent)
	case termbox.KeyArrowUp:
		if sv.sel > 0 {
			sv.sel--
		}
	case termbox.KeyArrowDown:
		if sv.sel+1 < len(sv.settings) {
			sv.sel++
		}
	case termbox.KeyEnter:
		sv.GiveFocus(sv.settings[sv.sel])
	}
}

func (sv *InputView) HandleRune(r rune) {
	if !sv.hasFocus {
		sv.settings[sv.sel].HandleRune(r)
		return
	}
}
