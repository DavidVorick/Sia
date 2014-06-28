package siafiles

import (
	"io"
	"os"
)

func Copy(src string, dst string) (err error) {
	s, err := os.Open(src)
	if err != nil {
		return
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return
}
