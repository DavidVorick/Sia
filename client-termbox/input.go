package main

import (
	"github.com/nsf/termbox-go"
)

const (
	ButtonColor   = termbox.ColorBlue
	ButtonHLColor = termbox.ColorRed

	CheckboxColor   = termbox.ColorDefault
	CheckboxHLColor = termbox.ColorRed

	FieldColor   = termbox.ColorBlue
	FieldHLColor = termbox.ColorRed
)

// An Input is simply a view that can be highlighted. This may cause some
// confusion, because Focus() appears to fill this role as well. However,
// implementation details make using Focus() for inputs ugly and unintuitive. I
// hope to find a better solution in the future.
type Input interface {
	View
	DrawHL()
}

// A Button is an Input that can trigger a function when pressed.
type Button struct {
	DefaultView
	label string
	press func()
}

func newButton(parent View, label string, press func()) *Button {
	b := &Button{
		label: " " + label + " ",
		press: press,
	}
	b.Parent = parent
	return b
}

func (b *Button) Draw() {
	drawColorString(b.MinX, b.MinY, b.label, termbox.ColorWhite, ButtonColor)
}

func (b *Button) DrawHL() {
	drawColorString(b.MinX, b.MinY, b.label, termbox.ColorWhite, ButtonHLColor)
}

// Buttons can only perform one action, so there is no need for them to have
// Focus. Accordingly, the Button immediately returns Focus to its parent after
// triggering the press() function. However, this means that the parent, not
// the Button, controls what user input triggers it. Whether this is a good
// idea remains to be seen.
func (b *Button) Focus() {
	b.hasFocus = true
	b.press()
	b.GiveFocus(b.Parent)
}

// A Checkbox is an Input that can be toggled on or off. The state of the
// Checkbox is tied to a boolean, which is supplied when the Checkbox is
// created.
type Checkbox struct {
	DefaultView
	label   string
	checked *bool
}

func newCheckbox(parent View, label string, checked *bool) *Checkbox {
	c := &Checkbox{
		label:   label,
		checked: checked,
	}
	c.Parent = parent
	return c
}

// Like a Button, a Checkbox can only perform one action. See the matching Button docstring.
func (c *Checkbox) Focus() {
	c.hasFocus = true
	*c.checked = !*c.checked
	c.GiveFocus(c.Parent)
}

func (c *Checkbox) Draw() {
	if *c.checked {
		drawColorString(c.MinX, c.MinY, "[X] "+c.label, termbox.ColorWhite, CheckboxColor)
	} else {
		drawColorString(c.MinX, c.MinY, "[ ] "+c.label, termbox.ColorWhite, CheckboxColor)
	}
}

func (c *Checkbox) DrawHL() {
	if *c.checked {
		drawColorString(c.MinX, c.MinY, "[X] "+c.label, termbox.ColorWhite, CheckboxHLColor)
	} else {
		drawColorString(c.MinX, c.MinY, "[ ] "+c.label, termbox.ColorWhite, CheckboxHLColor)
	}
}

// A Field is an Input that allows text entry. The text is tied to a string,
// which is supplied when the Field is created.
type Field struct {
	DefaultView
	text string
	ref  *string
	pos  int
}

// Focus places the cursor at the end of the text currently in the Field.
func (f *Field) Focus() {
	f.hasFocus = true
	f.pos = len(f.text)
	termbox.SetCursor(f.MinX+f.pos, f.MinY)
}

func (f *Field) Draw() {
	drawRectangle(f.Rectangle, FieldColor)
	drawColorString(f.MinX, f.MinY, f.text, termbox.ColorWhite, FieldColor)
}

func (f *Field) DrawHL() {
	drawRectangle(f.Rectangle, FieldHLColor)
	drawColorString(f.MinX, f.MinY, f.text, termbox.ColorWhite, FieldHLColor)
}

// Fields pose an implementation problem that Buttons and Checkboxes do not.
// Since the arrow keys are used to control the cursor in the Field, they
// cannot be used to navigate menus while the user is editing text. This means
// that an extra key is required to move in and out of "edit mode." It isn't a
// pretty solution, but it works for now.
func (f *Field) HandleKey(key termbox.Key) {
	switch key {
	case termbox.KeyEnter:
		// save current text to ref
		*f.ref = f.text
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
		f.HandleRune(' ')
	case termbox.KeyTab:
		f.HandleRune('\t')
	case termbox.KeyDelete:
		if f.pos < len(f.text) {
			f.deleteForward()
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if f.pos > 0 {
			f.deleteBackward()
			f.pos--
			f.updateCursor()
		}
	}
}

// HandleRune does not yet support unicode.
func (f *Field) HandleRune(r rune) {
	if len(f.text) >= f.MaxX-f.MinX-1 {
		return
	}
	f.text = f.text[:f.pos] + string(r) + f.text[f.pos:]
	f.pos++
	f.updateCursor()
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

// A Form is an Input that combines a Field with a label.
type Form struct {
	Rectangle
	Field
	label string
	width int
}

func newForm(parent View, label string, ref *string, width int) *Form {
	f := &Form{
		label: label,
		width: width,
	}
	f.ref = ref
	f.text = *ref
	f.Parent = parent
	return f
}

func (f *Form) SetDims(r Rectangle) {
	r.MaxX = r.MinX + len(f.label) + f.width
	r.MaxY = r.MinY + 1
	f.Rectangle = r

	r.MinX += len(f.label) + 1
	f.Field.SetDims(r)
}

func (f *Form) Draw() {
	drawString(f.MinX, f.MinY, f.label)
	f.Field.Draw()
}

func (f *Form) DrawHL() {
	drawString(f.MinX, f.MinY, f.label)
	f.Field.DrawHL()
}

// An InputsView is a collection of Inputs that can be navigated.
type InputsView struct {
	DefaultView
	inputs  []Input
	offsets []int
	sel     int
}

func (iv *InputsView) SetDims(r Rectangle) {
	iv.Rectangle = r
	for i := range iv.inputs {
		// inputs are fixed size, so they only care about MinX/MinY
		iv.inputs[i].SetDims(Rectangle{
			MinX: r.MinX + 1,
			MinY: r.MinY + iv.offsets[i],
		})
	}
}

func (iv *InputsView) Draw() {
	for i, in := range iv.inputs {
		if i == iv.sel && iv.hasFocus {
			in.DrawHL()
		} else {
			in.Draw()
		}
	}
}

func (sv *InputsView) HandleKey(key termbox.Key) {
	if !sv.hasFocus {
		sv.inputs[sv.sel].HandleKey(key)
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
		if sv.sel+1 < len(sv.inputs) {
			sv.sel++
		}
	case termbox.KeyEnter:
		sv.GiveFocus(sv.inputs[sv.sel])
	}
}

func (sv *InputsView) HandleRune(r rune) {
	if !sv.hasFocus {
		sv.inputs[sv.sel].HandleRune(r)
		return
	}
}
