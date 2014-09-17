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
	Rectangle
	Parent   View
	hasFocus bool
	title    string
	fields   []Field
}

func (s *Section) SetDims(r Rectangle) {
	s.Rectangle = r
}

func (s *Section) GiveFocus() {
	s.hasFocus = true
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

}

type SettingsView struct {
	Rectangle
	Parent   View
	hasFocus bool
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

func (sv *SettingsView) GiveFocus() {
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
		sv.hasFocus = false
		termbox.HideCursor()
		sv.Parent.GiveFocus()
	}
}

func newSettingsView(parent View) View {
	sv := &SettingsView{
		Parent:   parent,
		hasFocus: false,
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
	for i := range sv.sections {
		sv.sections[i].Parent = sv
	}
	return sv
}
