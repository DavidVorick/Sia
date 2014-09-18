package main

import (
	"github.com/nsf/termbox-go"
)

const (
	FieldColor = termbox.ColorBlue
)

type Field struct {
	DefaultView
	name   string
	width  int
	offset int
}

func (f *Field) Focus() {
	f.hasFocus = true
	// move cursor to first field
	termbox.SetCursor(f.MinX+1+len(f.name), f.MinY)
}

func (f *Field) Draw() {
	clearRectangle(f.Rectangle)
	drawString(f.MinX, f.MinY, f.name, termbox.ColorWhite, termbox.ColorDefault)
	drawLine(f.MinX+len(f.name)+1, f.MinY, f.width, termbox.ColorBlue)
}

func (f *Field) HandleKey(key termbox.Key) {
}

type SettingsView struct {
	DefaultView
	fields []*Field
	sel    int
}

func (sv *SettingsView) SetDims(r Rectangle) {
	sv.Rectangle = r
	for _, f := range sv.fields {
		f.SetDims(Rectangle{
			MinX: r.MinX + 1,
			MinY: r.MinY + f.offset,
			MaxX: r.MinX + len(f.name) + f.width + 2,
			MaxY: r.MinY + f.offset + 1,
		})
	}
}

func (sv *SettingsView) Focus() {
	// focus the first field
	sv.GiveFocus(sv.fields[0])
}

func (sv *SettingsView) Draw() {
	for i := range sv.fields {
		sv.fields[i].Draw()
	}
}

func (sv *SettingsView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		termbox.HideCursor()
		sv.GiveFocus(sv.Parent)
	case termbox.KeyArrowUp:
		if sv.sel > 0 {
			sv.sel--
		}
		sv.GiveFocus(sv.fields[sv.sel])
	case termbox.KeyArrowDown:
		if sv.sel+1 < len(sv.fields) {
			sv.sel++
		}
		sv.GiveFocus(sv.fields[sv.sel])
	}
}

func newSettingsView(parent View) View {
	sv := &SettingsView{
		fields: []*Field{
			{name: "Client Port:", width: 20, offset: 1},
			{name: "Server Host:", width: 20, offset: 3},
			{name: "Server Port:", width: 20, offset: 4},
			{name: "Server ID:  ", width: 20, offset: 5},
		},
	}
	sv.Parent = parent
	for i := range sv.fields {
		sv.fields[i].Parent = sv
	}
	return sv
}
