package vterm

import (
	"github.com/aaronduino/i3-tmux/cursor"
)

// Blinker handles blinking the Char under the cursor
type Blinker struct {
	X, Y    int
	Visible bool
}

// StartBlinker starts blinking a cursor at the vterm's selection.
// The cursor should be visible immediately after this function is called.
func (v *VTerm) StartBlinker() {
	v.Blinker.Visible = true
}

// updateBlinker sends the char under the cursor to the renderer to keep the host's cursor location up-to-date
func (v *VTerm) updateBlinker() {
	if len(v.screen) > v.Cursor.Y && len(v.screen[v.Cursor.Y]) > v.Cursor.X {
		char := v.screen[v.Cursor.Y][v.Cursor.X]
		char.Cursor.X = v.Cursor.X
		char.Cursor.Y = v.Cursor.Y
		v.out <- char
	} else {
		v.out <- Char{
			Rune: ' ',
			Cursor: cursor.Cursor{
				X: v.Cursor.X, Y: v.Cursor.Y,
			},
		}
	}
}

// StopBlinker immediately hides and stops blinking the vterm's cursor.
func (v *VTerm) StopBlinker() {
	v.Blinker.Visible = false
}
