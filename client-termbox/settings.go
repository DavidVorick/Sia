package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	SettingColor      = termbox.ColorBlue
	SettingFocusColor = termbox.ColorRed
)

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

type SettingsView struct {
	DefaultView
	settings []*Setting
	sel      int
}

func (sv *SettingsView) SetDims(r Rectangle) {
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

func (sv *SettingsView) Draw() {
	for i, s := range sv.settings {
		if i == sv.sel {
			s.SetColor(SettingFocusColor)
		} else {
			s.SetColor(SettingColor)
		}
		s.Draw()
	}
}

func (sv *SettingsView) HandleKey(key termbox.Key) {
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

func (sv *SettingsView) HandleChar(r rune) {
	if !sv.hasFocus {
		sv.settings[sv.sel].HandleChar(r)
		return
	}
}

func newSettingsView(parent View) View {
	// convert config values to strings
	clientPort := fmt.Sprint(config.Client.Port)
	serverPort := fmt.Sprint(config.Server.Port)
	serverID := fmt.Sprint(config.Server.ID)

	sv := &SettingsView{
		settings: []*Setting{
			{Field: Field{text: clientPort}, name: "Client Port:", width: 20, offset: 1},
			{Field: Field{text: config.Server.Host}, name: "Server Host:", width: 20, offset: 3},
			{Field: Field{text: serverPort}, name: "Server Port:", width: 20, offset: 4},
			{Field: Field{text: serverID}, name: "Server ID:  ", width: 20, offset: 5},
		},
	}
	sv.Parent = parent
	for _, s := range sv.settings {
		s.Parent = sv
		s.color = SettingColor
	}
	return sv
}
