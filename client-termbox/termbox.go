package main

import (
	"github.com/nsf/termbox-go"
)

type Rectangle struct {
	MinX, MinY, MaxX, MaxY int
}

func drawString(x, y int, s string, fg, bg termbox.Attribute) {
	for i, c := range s {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}
