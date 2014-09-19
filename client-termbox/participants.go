package main

import (
	"github.com/nsf/termbox-go"
)

const (
	ParticipantMenuWidth = 15
)

type ParticipantMenuView struct {
	MenuView
}

func (pmv *ParticipantMenuView) Focus() {
	pmv.hasFocus = true
	pmv.loadParticipants()
}

func newParticipantMenuView(parent View) View {
	pmv := new(ParticipantMenuView)
	pmv.Parent = parent
	pmv.Title = "Participants"
	pmv.MenuWidth = ParticipantMenuWidth
	// load participant names and create views
	pmv.loadParticipants()
	return pmv
}

func (pmv *ParticipantMenuView) loadParticipants() {
	names, err := server.GetParticipantNames()
	if err != nil {
		//drawError("Could not load participants:", err)
		pmv.Items = []string{"<empty>"}
		pmv.Windows = []View{&DefaultView{Parent: pmv}}
		return
	}
	for _, n := range names {
		pmv.addParticipant(n)
	}
}

func (pmv *ParticipantMenuView) addParticipant(name string) {
	// create participant view
	pv := new(ParticipantView)
	pv.Parent = pmv
	pv.name = name

	pmv.Items = append(pmv.Items, name)
	pmv.Windows = append(pmv.Windows, pv)
}

type ParticipantView struct {
	DefaultView
	name string
}

func (pv *ParticipantView) Draw() {

}

func (pv *ParticipantView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		pv.GiveFocus(pv.Parent)
	}
}
