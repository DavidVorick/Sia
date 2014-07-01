package siafiles

import (
	"io"
)

type DoubleReader struct {
	a io.Reader
	b io.Reader
}

func NewDoubleReader(a io.Reader, b io.Reader) (dr *DoubleReader) {
	dr = new(DoubleReader)
	dr.a = a
	dr.b = b
	return
}

// Read on a double reader first reads from the reader a. If a does not fill
// the entire byte slice, the remainder of the slice is filled by b.
func (dr *DoubleReader) Read(b []byte) (n int, err error) {
	var m int
	n, _ = dr.a.Read(b)
	if n != len(b) {
		m, err = dr.b.Read(b[n:])
	}
	n += m
	return
}
