package siafiles

import (
	"bytes"
	"os"
)

type scratch struct {
	bytes.Buffer
}

// Scratch creates a new scratch. A scratch is a temporary buffer whose
// contents are only saved to disk when SaveAs() is called.
func Scratch() *scratch {
	return &scratch{}
}

// SaveAs copies the contents of a scratch file into a permanent location on disk.
func (s *scratch) SaveAs(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	_, err = s.WriteTo(f)
	return
}

// PadTo pads the scratch with zeros. If length is less than the current length
// of the buffer, PadTo has no effect.
func (s *scratch) PadTo(length int) {
	if length < s.Len() {
		return
	}
	s.Write(make([]byte, length-s.Len()))
}
