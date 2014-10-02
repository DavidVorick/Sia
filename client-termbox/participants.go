package main

import (
	"strconv"

	"github.com/nsf/termbox-go"
)

const (
	ParticipantMenuWidth = 18
)

// ParticipantMenuMVC lists the Participants available to the user, and allows
// for the creation of new Participants.
type ParticipantMenuMVC struct {
	MenuMVC
}

func newParticipantMenuMVC(parent MVC) *ParticipantMenuMVC {
	pm := new(ParticipantMenuMVC)
	pm.Parent = parent
	pm.Title = "Participants"
	pm.MenuWidth = ParticipantMenuWidth
	pm.Items = []string{"New Participant"}
	pm.Windows = []MVC{newParticipantCreator(pm)}
	// load participant names and create views
	pm.loadParticipants()
	return pm
}

func (pm *ParticipantMenuMVC) Focus() {
	//pm.loadParticipants()
	pm.MenuMVC.Focus()
}

func (pm *ParticipantMenuMVC) loadParticipants() {
	names, err := server.GetParticipantNames()
	if err != nil {
		drawError("Could not load participants:", err)
		return
	}
	for _, n := range names {
		pm.addParticipant(n)
	}
}

// TODO: same as WalletMenuMVC.addWallet
func (pm *ParticipantMenuMVC) addParticipant(name string) {
	pv := new(ParticipantMVC)
	pv.Parent = pm
	pv.name = name

	pm.Items = append(pm.Items, name)
	pm.Windows = append(pm.Windows, pv)
}

// A ParticipantMVC displays the properties of a Participant.
type ParticipantMVC struct {
	DefaultMVC
	name string
}

func (p *ParticipantMVC) Draw() {
	// display properities of participant
}

func (p *ParticipantMVC) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyArrowLeft:
		p.GiveFocus(p.Parent)
	}
}

// The ParticipantCreator allows for the creation of new Participants.
type ParticipantCreator struct {
	InputGroupMVC
	name      string
	siblingID string
	customDir string
	bootstrap bool
}

func newParticipantCreator(parent MVC) *ParticipantCreator {
	pc := new(ParticipantCreator)
	pc.inputs = []Input{
		newForm(pc, "Name:      ", &pc.name, 20),
		newForm(pc, "Sibling ID:", &pc.siblingID, 20),
		newForm(pc, "Custom Dir:", &pc.customDir, 20),
		newCheckbox(pc, "New quorum", &pc.bootstrap),
		newButton(pc, "Create", pc.create),
	}
	pc.offsets = []int{1, 2, 3, 4, 6}
	pc.Parent = parent
	return pc
}

func (pc *ParticipantCreator) create() {
	// validate values
	if pc.name == "" {
		drawError("Please provide a name")
		return
	}
	id, err := strconv.ParseUint(pc.siblingID, 10, 64)
	if err != nil {
		drawError("Invalid sibling ID")
		return
	}

	err = server.CreateParticipant(pc.name, id, pc.customDir, pc.bootstrap)
	if err != nil {
		drawError("Participant creation failed:", err)
	} else {
		drawInfo("Created " + pc.name)
	}
}
