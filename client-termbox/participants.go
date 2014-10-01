package main

import (
	"github.com/nsf/termbox-go"
)

const (
	ParticipantMenuWidth = 18
)

type ParticipantMenuView struct {
	MenuView
}

func newParticipantMenuView(parent View) *ParticipantMenuView {
	pmv := new(ParticipantMenuView)
	pmv.Parent = parent
	pmv.Title = "Participants"
	pmv.MenuWidth = ParticipantMenuWidth
	pmv.Items = []string{"New Participant"}
	pmv.Windows = []View{newParticipantCreator(pmv)}
	// load participant names and create views
	pmv.loadParticipants()
	return pmv
}

func (pmv *ParticipantMenuView) Focus() {
	//pmv.loadParticipants()
	pmv.MenuView.Focus()
}

func (pmv *ParticipantMenuView) loadParticipants() {
	names, err := server.GetParticipantNames()
	if err != nil {
		//drawError("Could not load participants:", err)
		return
	}
	for _, n := range names {
		pmv.addParticipant(n)
	}
}

func (pmv *ParticipantMenuView) addParticipant(name string) {
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
	// display properities of participant
}

func (pv *ParticipantView) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		pv.GiveFocus(pv.Parent)
	}
}

type ParticipantCreator struct {
	InputsView
	name      string
	siblingID string
	customDir string
	genesis   bool
}

func newParticipantCreator(parent View) *ParticipantCreator {
	pc := new(ParticipantCreator)
	pc.inputs = []Input{
		newForm(pc, "Name:      ", &pc.name, 20, 1),
		newForm(pc, "Sibling ID:", &pc.siblingID, 20, 2),
		newForm(pc, "Custom Dir:", &pc.customDir, 20, 3),
		newCheckbox(pc, "Genesis", &pc.genesis, 4),
		newButton(pc, "Submit", pc.create, 6),
	}
	pc.Parent = parent
	return pc
}

func (pc *ParticipantCreator) create() {

}
