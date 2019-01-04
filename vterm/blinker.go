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
		oldChar := v.buffer[v.blinker.y][v.blinker.x]
		if oldChar.Rune == 0 {
			oldChar.Rune = ' '
		}
		v.out <- oldChar
	}

	cleanUp()

	if v.cursor.Y > len(v.buffer)-1 || v.cursor.X > len(v.buffer[v.cursor.Y])-1 {
		return
	}

	v.blinker.x = v.cursor.X
	v.blinker.y = v.cursor.Y

	char := v.buffer[v.blinker.y][v.blinker.x]
	char.Cursor.Bg = cursor.Color{
		ColorMode: cursor.ColorBit3Bright,
		Code:      7,
	}
	if char.Rune == 0 {
		char.Rune = ' '
	}
	// char.Rune = '0'
	v.out <- char

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
