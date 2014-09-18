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
		f.updateCursor()
	case termbox.KeyArrowRight:
		if f.pos < len(f.text) {
			f.pos++
		}
		f.updateCursor()
	case termbox.KeySpace:
		f.HandleChar(' ')
		f.Draw()
	case termbox.KeyTab:
		f.HandleChar('\t')
		f.Draw()
	case termbox.KeyDelete:
		if f.pos < len(f.text) {
			f.deleteForward()
		}
		f.Draw()
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if f.pos > 0 {
			f.deleteBackward()
			f.pos--
			f.updateCursor()
		}
		f.Draw()
	}
}

func (f *Field) HandleChar(r rune) {
	if len(f.text) >= f.MaxX-f.MinX-1 {
		return
	}
	f.text = f.text[:f.pos] + string(r) + f.text[f.pos:]
	f.pos++
	f.updateCursor()
	f.Draw()
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
