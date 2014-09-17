package main

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

type Rectangle struct {
	MinX, MinY, MaxX, MaxY int
}

func drawRectangle(r Rectangle, color termbox.Attribute) {
	for x := r.MinX; x < r.MaxX; x++ {
		for y := r.MinY; y < r.MaxY; y++ {
			termbox.SetCell(x, y, ' ', color, color)
		}
	}
}

func clearRectangle(r Rectangle) {
	drawRectangle(r, termbox.ColorDefault)
}

func drawString(x, y int, s string, fg, bg termbox.Attribute) {
	for i, c := range s {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

func drawError(v ...interface{}) {
	s := strings.Trim(fmt.Sprintln(v...), "\n")
	strings.Trim(s, s)
	w, h := termbox.Size()
	drawRectangle(Rectangle{0, h - 1, w, h}, termbox.ColorRed)
	drawString(1, h-1, s, termbox.ColorWhite, termbox.ColorRed)
}
