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

type Input interface {
	View
	DrawHL()
}

type Button struct {
	DefaultView
	label  string
	offset int
	press  func()
}

func newButton(parent View, label string, press func(), offset int) *Button {
	b := &Button{
		label:  " " + label + " ",
		offset: offset,
		press:  press,
	}
	b.Parent = parent
	return b
}

func (b *Button) SetDims(r Rectangle) {
	r.MinY += b.offset
	r.MaxY += b.offset
	b.Rectangle = r
}

func (b *Button) Draw() {
	drawColorString(b.MinX, b.MinY, b.label, termbox.ColorWhite, ButtonColor)
}

func (b *Button) DrawHL() {
	drawColorString(b.MinX, b.MinY, b.label, termbox.ColorWhite, ButtonHLColor)
}

func (b *Button) Focus() {
	b.hasFocus = true
	b.press()
	b.GiveFocus(b.Parent)
}

type Checkbox struct {
	DefaultView
	label   string
	offset  int
	checked *bool
}

func newCheckbox(parent View, label string, checked *bool, offset int) *Checkbox {
	c := &Checkbox{
		label:   label,
		offset:  offset,
		checked: checked,
	}
	c.Parent = parent
	return c
}

func (c *Checkbox) SetDims(r Rectangle) {
	r.MinY += c.offset
	r.MaxY += c.offset
	c.Rectangle = r
}

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

type Field struct {
	DefaultView
	text string
	ref  *string
	pos  int
}

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

type Form struct {
	Rectangle
	Field
	label  string
	width  int
	offset int
}

func newForm(parent View, label string, ref *string, width, offset int) *Form {
	f := &Form{
		label:  label,
		width:  width,
		offset: offset,
	}
	f.ref = ref
	f.text = *ref
	f.Parent = parent
	return f
}

func (f *Form) SetDims(r Rectangle) {
	r.MinY += f.offset
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

type InputsView struct {
	DefaultView
	inputs []Input
	sel    int
}

func (iv *InputsView) SetDims(r Rectangle) {
	iv.Rectangle = r
	for _, i := range iv.inputs {
		i.SetDims(Rectangle{
			MinX: r.MinX + 1,
			MinY: r.MinY,
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
