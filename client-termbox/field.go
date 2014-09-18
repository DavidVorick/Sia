package main

import (
	"github.com/nsf/termbox-go"
)

const (
	FieldFocusColor = termbox.ColorRed
)

type Field struct {
	DefaultView
	color termbox.Attribute
	text  string
}

func (f *Field) Focus() {
	f.hasFocus = true
	termbox.SetCursor(f.MinX+len(f.text), f.MinY)
}

func (f *Field) Draw() {
	drawRectangle(f.Rectangle, f.color)
	drawString(f.MinX, f.MinY, f.text, termbox.ColorWhite, f.color)
}

func (f *Field) HandleKey(key termbox.Key) {
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
