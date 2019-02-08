/*
Package keypress is a library for advanced keypress detection and parsing.
*/
package keypress

import (
	gc "github.com/rthornton128/goncurses"
)

var win *gc.Window
var callback func(name string, raw []byte)

func next() gc.Key {
	return win.GetChar()
}

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(w *gc.Window, c func(name string, raw []byte)) {
	win = w
	callback = c

	for {
		switch key := win.GetChar(); key {
		case 23: // Ctrl+W
			return
		case 27:
			handleEscapeCode()
		default:
			callback("", []byte{byte(key)})
		}
	}
}
