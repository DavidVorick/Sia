package main

import (
	"github.com/nsf/termbox-go"
)

const (
	SectionColor = termbox.ColorWhite
	FieldColor   = termbox.ColorBlue
)

type Field struct {
	name  string
	width int
}

type Section struct {
	DefaultView
	title  string
	fields []Field
}

func (s *Section) Draw() {
	clearRectangle(s.Rectangle)
	drawString(s.MinX+1, s.MinY+1, s.title, SectionColor, termbox.ColorDefault)
	for i, f := range s.fields {
		drawString(s.MinX+5, s.MinY+i+2, f.name, SectionColor, termbox.ColorDefault)
		drawLine(s.MinX+6+len(f.name), s.MinY+i+2, f.width+1, FieldColor)
	}
}

func (s *Section) HandleInput(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		s.GiveFocus(s.Parent)
	}
}

type SettingsView struct {
	DefaultView
	sections []Section
	sel      int
}

func (sv *SettingsView) SetDims(r Rectangle) {
	sv.Rectangle = r
	for i := range sv.sections {
		sv.sections[i].SetDims(r)
		r.MinY += len(sv.sections[i].fields)*2 + 1
	}
}

func (sv *SettingsView) Focus() {
	sv.hasFocus = true
	// move cursor to first field
	termbox.SetCursor(sv.MinX+6+len(sv.sections[0].fields[0].name), sv.MinY+2)
}

func (sv *SettingsView) Draw() {
	for i := range sv.sections {
		sv.sections[i].Draw()
	}
}

func (sv *SettingsView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		termbox.HideCursor()
		sv.GiveFocus(sv.Parent)
	}
}

func newSettingsView(parent View) View {
	sv := &SettingsView{
		sections: []Section{
			{title: "Client", fields: []Field{
				{name: "Port:", width: 20},
			}},
			{title: "Server", fields: []Field{
				{name: "Host:", width: 20},
				{name: "Port:", width: 20},
				{name: "ID:  ", width: 20},
			}},
		},
	}
	sv.Parent = parent
	for i := range sv.sections {
		sv.sections[i].Parent = sv
	}
	return sv
}
