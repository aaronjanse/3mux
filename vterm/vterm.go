/*
Package vterm provides a layer of abstraction between a channel of incoming text (possibly containing ANSI escape codes, et al) and a channel of outbound Char's.

A Char is a character printed using a given cursor (which is stored alongside the Char).
*/
package vterm

import (
	"sync"

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

	buffer      [][]Char
	bufferMutux *sync.Mutex

	cursor cursor.Cursor

	in  <-chan rune
	out chan<- Char

	storedCursorX, storedCursorY int

	blinker *Blinker
}

// NewVTerm returns a VTerm ready to be used by its exported methods
func NewVTerm(in <-chan rune, out chan<- Char) *VTerm {
	w := 30
	h := 10

	buffer := [][]Char{}
	for j := 0; j < h; j++ {
		row := []Char{}
		for i := 0; i < w; i++ {
			row = append(row, Char{
				Rune:   ' ',
				Cursor: cursor.Cursor{X: i, Y: j},
			})
		}
		buffer = append(buffer, row)
	}

	return &VTerm{
		w:           w,
		h:           h,
		buffer:      buffer,
		bufferMutux: &sync.Mutex{},
		cursor:      cursor.Cursor{X: 0, Y: 0},
		in:          in,
		out:         out,
		blinker:     newBlinker(),
	}
}

// Reshape safely updates a VTerm's width & height
func (v *VTerm) Reshape(w, h int) {
	v.bufferMutux.Lock()
	v.w = w
	v.h = h

	// // clear the relevant area of the screen
	// for j := 0; j < v.h; j++ {
	// 	for i := 0; i < v.w; i++ {
	// 		v.out <- Char{
	// 			Rune:   '_',
	// 			Cursor: cursor.Cursor{X: i, Y: j},
	// 		}
	// 	}
	// }

	v.bufferMutux.Unlock()
}

// RedrawWindow draws the entire visible window from scratch, sending the Char's to the scheduler via the out channel
func (v *VTerm) RedrawWindow() {
	v.bufferMutux.Lock()

	var visibleBuffer [][]Char
	if len(v.buffer) > v.h {
		visibleBuffer = v.buffer[len(v.buffer)-v.h:]
	} else {
		visibleBuffer = v.buffer
	}

	for j := 0; j <= v.h; j++ {
		var row []Char
		if j < len(visibleBuffer) {
			row = visibleBuffer[j]
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

	v.bufferMutux.Unlock()
}
