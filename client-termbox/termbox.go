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
	// Note that if MinY == MaxY, nothing is drawn. Usually this means you
	// should be calling drawLine instead.
	for x := r.MinX; x < r.MaxX; x++ {
		for y := r.MinY; y < r.MaxY; y++ {
			termbox.SetCell(x, y, ' ', color, color)
		}
	}
}

func drawLine(x, y, w int, color termbox.Attribute) {
	for i := x; i < x+w; i++ {
		termbox.SetCell(i, y, ' ', color, color)
	}
}

func drawColorString(x, y int, s string, fg, bg termbox.Attribute) {
	// i must be manually incremented because range iterates over code points,
	// not bytes, meaning i would be incremented multiple times per rune.
	var i int
	for _, c := range s {
		termbox.SetCell(x+i, y, c, fg, bg)
		i++
	}
}

func drawString(x, y int, s string) {
	drawColorString(x, y, s, termbox.ColorWhite, termbox.ColorDefault)
}

func drawError(v ...interface{}) {
	s := strings.Trim(fmt.Sprintln(v...), "\n")
	strings.Trim(s, s)
	w, h := termbox.Size()
	drawRectangle(Rectangle{0, h - 1, w, h}, termbox.ColorRed)
	drawColorString(1, h-1, s, termbox.ColorWhite, termbox.ColorRed)
}
