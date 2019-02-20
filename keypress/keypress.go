/*
Package keypress is a library for advanced keypress detection and parsing.

It provides a callback with both a human-readable name for the event detected, along with the raw data behind it.

The following are supported human-readable names:
	Enter
	Esc

	Mouse Down
	Mouse Up

	Scroll Up
	Scroll Down

	[letter]
	Alt+[letter]
	Alt+Shift+[letter]
	Ctrl+[letter]

	Up
	Down
	Left
	Right

	Shift+[arrow]
	Alt+[arrow]
	Alt+Shift+[arrow]
	Ctrl+[arrow]
*/
package keypress

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	term "github.com/nsf/termbox-go"
)

const debugKeycodes = true

var directionNames = map[byte]string{
	65: "Up",
	66: "Down",
	67: "Right",
	68: "Left",
}

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(callback func(name string, data []byte)) {
	err := term.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// show cursor, make it blink
	fmt.Print("\033[?25h\033[?12h") // EXPLAIN: do we need this?

	for {
		raw := make([]byte, 16) // EXPLAIN: why 16?
		ev := term.PollRawEvent(raw)

		data := raw[:ev.N]

		handle := func(name string) {
			callback(name, data)
		}

		switch data[0] {
		case 13:
			handle("Enter")
		case 195: // Alt
			parseAltLetter(data[1]-64, handle)
		case 27: // Escape code
			handleEscapeCode(data, handle)
		default:
			if len(data) == 1 {
				if data[0] <= 26 { // Ctrl
					letter := string('A' + data[0] - 1)
					if letter == "Q" { // exit upon Ctrl+Q
						return
					}
					handle("Ctrl+" + letter)
				} else {
					letter := string(data[0])
					handle(letter)
				}
			}
		}

		if debugKeycodes {
			log.Println(ev)
			log.Print(data)
			str := ""
			for _, b := range data {
				str += string(b)
			}
			log.Println(str)
			log.Println()
		}
	}
}

func handleEscapeCode(data []byte, handle func(name string)) {
	if len(data) == 1 { // Lone Esc Key
		handle("Esc")
		return
	}

	switch data[1] {
	case 79:
		direction := directionNames[data[2]]
		if len(data) == 15 {
			handle("Scroll " + direction)
		} else {
			handle(direction)
		}
	case 91:
		switch data[2] {
		case 51: // Mouse
			code := string(data[2:])
			code = strings.TrimSuffix(code, "M") // NOTE: are there other codes we are forgetting about?
			pieces := strings.Split(code, ";")
			switch pieces[0] {
			case "32":
				handle("Mouse Down")
			case "35":
				handle("Mouse Up")
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		case 57: // Scrolling
			switch data[3] {
			case 54:
				handle("Scroll Up")
			case 55:
				handle("Scroll Down")
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		default:
			arrow := directionNames[data[5]]
			switch data[4] {
			case 50:
				handle("Shift+" + arrow)
			case 51:
				handle("Alt+" + arrow)
			case 52:
				handle("Alt+Shift+" + arrow)
			case 53:
				handle("Ctrl+" + arrow)
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		}
	default:
		parseAltLetter(data[1], handle)
	}
}

func parseAltLetter(b byte, handle func(name string)) {
	letter := rune(b)
	if unicode.IsUpper(letter) {
		handle("Alt+Shift+" + string(letter))
	} else {
		handle("Alt+" + string(unicode.ToUpper(letter)))
	}
}
