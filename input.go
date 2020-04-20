package main

import (
	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/wm"
)

// handleInput puts the input through a series of switches and seive functions.
// When something acts on the event, we stop passing it downstream
func handleInput(u *wm.Universe, human string, obj ecma48.Output) {
	if seiveMouseEvents(u, human, obj) {
		return
	}

	if seiveConfigEvents(u, human) {
		return
	}

	// if we didn't find anything special, just pass the raw data to the selected terminal
	u.HandleStdin(obj)
}

var mouseDownX, mouseDownY int

// seiveMouseEvents processes mouse events and returns true if the data should *not* be passed downstream
func seiveMouseEvents(u *wm.Universe, human string, obj ecma48.Output) bool {
	switch ev := obj.Parsed.(type) {
	case ecma48.MouseDown:
		u.SelectAtCoords(ev.X, ev.Y)
		mouseDownX = ev.X
		mouseDownY = ev.Y
	case ecma48.MouseUp:
		u.DragBorder(mouseDownX, mouseDownY, ev.X, ev.Y)
	case ecma48.MouseDrag:
		// do nothing
	case ecma48.ScrollUp:
		u.ScrollUp()
	case ecma48.ScrollDown:
		u.ScrollDown()
	default:
		return false
	}

	return true
}
