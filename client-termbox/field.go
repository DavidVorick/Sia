package main

import (
	"github.com/nsf/termbox-go"
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
		termbox.SetCursor(f.MinX+f.pos, f.MinY)
	case termbox.KeyArrowRight:
		if f.pos < len(f.text) {
			f.pos++
		}
		termbox.SetCursor(f.MinX+f.pos, f.MinY)
	}
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
