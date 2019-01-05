package vterm

import (
	"time"

	"github.com/aaronduino/i3-tmux/cursor"
)

// Blinker handles blinking the Char under the cursor
type Blinker struct {
	x, y         int
	fadeDone     chan bool
	blinking     bool
	stopBlinking chan bool
}

func newBlinker() *Blinker {
	return &Blinker{
		x:            0,
		y:            0,
		blinking:     false,
		fadeDone:     make(chan bool),
		stopBlinking: make(chan bool, 2),
	}
}

// StartBlinker starts blinking a cursor at the vterm's selection.
// The cursor should be visible immediately after this function is called.
func (v *VTerm) StartBlinker() {
	v.startFade()

	v.blinker.blinking = true

	go (func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				v.startFade()
			case <-v.blinker.stopBlinking:
				// v.blinker.fadeDone <- true
				ticker.Stop()
				v.blinker.blinking = false
				return
			}
		}
	})()
}

func (v *VTerm) updateBlinker() {
	// v.blinker.fadeDone <- true

	// v.startFade()
}

func (v *VTerm) startFade() {
	cleanUp := func() {
		if v.blinker.y > len(v.screen)-1 || v.blinker.x > len(v.screen[v.blinker.y])-1 {
			char := Char{
				Rune: ' ',
				Cursor: cursor.Cursor{
					X: v.blinker.x, Y: v.blinker.y,
				},
			}
			v.out <- char
		} else {
			oldChar := v.screen[v.blinker.y][v.blinker.x]
			if oldChar.Rune == 0 {
				oldChar.Rune = ' '
			}
			oldChar.Cursor.X = v.blinker.x
			oldChar.Cursor.Y = v.blinker.y
			v.out <- oldChar
		}
	}

	cleanUp()

	v.blinker.x = v.cursor.X
	v.blinker.y = v.cursor.Y

	if v.cursor.Y > len(v.screen)-1 || v.cursor.X > len(v.screen[v.cursor.Y])-1 {
		char := Char{
			Rune: ' ',
			Cursor: cursor.Cursor{
				X: v.blinker.x, Y: v.blinker.y,
				Bg: cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      7,
				},
			},
		}
		v.out <- char
	} else {
		char := v.screen[v.blinker.y][v.blinker.x]
		char.Cursor.Bg = cursor.Color{
			ColorMode: cursor.ColorBit3Bright,
			Code:      7,
		}
		char.Cursor.X = v.blinker.x
		char.Cursor.Y = v.blinker.y
		if char.Rune == 0 {
			char.Rune = ' '
		}
		v.out <- char
	}

	go (func() {
		fadeTimer := time.NewTimer(time.Second / 2)
		select {
		case <-v.blinker.fadeDone:
			cleanUp()
			fadeTimer.Stop()
			return
		case <-fadeTimer.C:
			cleanUp()
			return
		}
	})()
}

// StopBlinker immediately hides and stops blinking the vterm's cursor.
func (v *VTerm) StopBlinker() {
	if v.blinker.blinking {
		v.blinker.stopBlinking <- true
	}
}
