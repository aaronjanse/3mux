/*
Package vterm provides a layer of abstraction between a channel of incoming text (possibly containing ANSI escape codes, et al) and a channel of outbound Char's.

A Char is a character printed using a given cursor (which is stored alongside the Char).
*/
package vterm

import (
	"github.com/aaronduino/i3-tmux/cursor"
)

// Char represents one character in the terminal's grid
type Char struct {
	Rune rune

	Cursor cursor.Cursor
}

/*
VTerm acts as a virtual terminal emulator between a shell and the host terminal emulator

It both transforms an inbound stream of bytes into Char's and provides the option of dumping all the Char's that need to be rendered to display the currently visible terminal window from scratch.
*/
type VTerm struct {
	w, h int

	// visible screen; char cursor coords are ignored
	screen [][]Char

	scrollback [][]Char // disabled when using alt screen; char cursor coords are ignored

	usingAltScreen bool

	cursor cursor.Cursor

	in  <-chan rune
	out chan<- Char

	storedCursorX, storedCursorY int

	blinker *Blinker
}

// NewVTerm returns a VTerm ready to be used by its exported methods
func NewVTerm(in <-chan rune, out chan<- Char) *VTerm {
	w := 10
	h := 10

	screen := [][]Char{}
	for j := 0; j < h; j++ {
		row := []Char{}
		for i := 0; i < w; i++ {
			row = append(row, Char{
				Rune:   ' ',
				Cursor: cursor.Cursor{X: i, Y: j},
			})
		}
		screen = append(screen, row)
	}

	return &VTerm{
		w:              w,
		h:              h,
		screen:         screen,
		scrollback:     [][]Char{},
		usingAltScreen: false,
		cursor:         cursor.Cursor{X: 0, Y: 0},
		in:             in,
		out:            out,
		blinker:        newBlinker(),
	}
}

// Reshape safely updates a VTerm's width & height
func (v *VTerm) Reshape(w, h int) {
	if h > len(v.screen) { // move lines from scrollback
		linesToAdd := h - len(v.screen)
		if linesToAdd > len(v.scrollback) {
			linesToAdd = len(v.scrollback)
		}

		v.screen = append(v.scrollback[len(v.scrollback)-linesToAdd:], v.screen...)
		v.scrollback = v.scrollback[:len(v.scrollback)-linesToAdd]
	} else if h < len(v.screen) { // move lines to scrollback
		linesToMove := len(v.screen) - h

		v.scrollback = append(v.scrollback, v.screen[:linesToMove]...)
		v.screen = v.screen[linesToMove:]
	}

	v.w = w
	v.h = h
}

// RedrawWindow draws the entire visible window from scratch, sending the Char's to the scheduler via the out channel
func (v *VTerm) RedrawWindow() {
	for j := 0; j <= v.h; j++ {
		var row []Char
		if j < len(v.screen) {
			row = v.screen[j]
		} else {
			row = []Char{}
		}

		for i := 0; i <= v.w; i++ {
			if i < len(row) {
				char := row[i]
				char.Cursor.X = i
				char.Cursor.Y = j
				if char.Rune != 0 {
					v.out <- char
					continue
				}
			}
			v.out <- Char{
				Rune:   ' ',
				Cursor: cursor.Cursor{X: i, Y: j},
			}
		}
	}
}

func (v *VTerm) updateCursor() {
	if v.cursor.Y > v.h-1 { // move lines to scrollback
		linesToMove := v.cursor.Y - v.h + 1
		v.scrollback = append(v.scrollback, v.screen[:linesToMove]...)
		v.screen = v.screen[linesToMove:]

		v.cursor.Y = v.h - 1

		v.RedrawWindow()
	}
	v.updateBlinker()
}
