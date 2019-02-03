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

func (v *VTerm) updateBlinker() {
	if len(v.screen) > v.Cursor.Y && len(v.screen[v.Cursor.Y]) > v.Cursor.X {
		v.out <- v.screen[v.Cursor.Y][v.Cursor.X]
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
