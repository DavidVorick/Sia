package main

import (
	"fmt"
	"strconv"

	"github.com/NebulousLabs/Sia/state"

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
	pm.loadParticipants()
	pm.MenuMVC.Focus()
}

func (pm *ParticipantMenuMVC) loadParticipants() {
	names, err := server.ParticipantNames()
	if err != nil {
		drawError("Could not load participants:", err)
		return
	}

	// clear existing participants
	// TODO: only update participants that are new or have been removed
	pm.Items = pm.Items[:1]
	pm.Windows = pm.Windows[:1]

	for _, n := range names {
		pm.addParticipant(n)
	}
	// set dimensions of children
	pm.SetDims(pm.Rectangle)
}

func (pm *ParticipantMenuMVC) addParticipant(name string) {
	p := new(ParticipantMVC)
	p.Parent = pm
	p.name = name
	if err := server.ParticipantMetadata(name, &p.metadata); err != nil {
		drawError("Could not fetch metadata of "+name+":", err)
		return
	}
	pm.Items = append(pm.Items, name)
	pm.Windows = append(pm.Windows, p)
}

// A ParticipantMVC displays the properties of a Participant.
type ParticipantMVC struct {
	DefaultMVC
	name     string
	metadata state.Metadata
}

func (p *ParticipantMVC) Draw() {
	drawString(p.MinX+1, p.MinY+1, "Siblings:")
	for i, sib := range p.metadata.Siblings {
		var status string
		var color termbox.Attribute
		switch {
		case sib.Active():
			status, color = "Active", termbox.ColorGreen
		case sib.Inactive():
			status, color = "Inactive", termbox.ColorRed
		default:
			status = fmt.Sprintf("Passive for %d more compiles", sib.Status)
			color = termbox.ColorBlack | termbox.AttrBold
		}
		str := fmt.Sprintf("Sibling %d:", i)
		drawString(p.MinX+5, p.MinY+i+3, str)
		drawColorString(p.MinX+6+len(str), p.MinY+i+3, status, color, termbox.ColorDefault)
	}
	drawString(p.MinX+1, p.MinY+len(p.metadata.Siblings)+4, fmt.Sprintf("Height: %d", p.metadata.Height))
	drawString(p.MinX+1, p.MinY+len(p.metadata.Siblings)+5, fmt.Sprintf("Parent: %v", p.metadata.ParentBlock))
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
		return
	}

	drawInfo("Created " + pc.name)
	// TODO: need a good way of calling both:
	// ParticipantMenuMVC.loadParticipants() and
	// WalletMenuMVC.loadWallets()
	// This problem will only get worse as more features are added.
	// Maybe use channels?
}
