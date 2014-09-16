package main

import (
	"github.com/nsf/termbox-go"
)

type Rectangle struct {
	MinX, MinY, MaxX, MaxY int
}

func clearRectangle(r Rectangle) {
	for x := r.MinX; x < r.MaxX; x++ {
		for y := r.MinY; y < r.MaxY; y++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func drawString(x, y int, s string, fg, bg termbox.Attribute) {
	for i, c := range s {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}
