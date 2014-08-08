package siafiles

import (
	"os"
)

func Size(filename string) (size int64, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return
	}
	size = info.Size()
	return
}
