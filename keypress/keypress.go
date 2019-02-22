/*
Package keypress is a library for advanced keypress detection and parsing.

It provides a callback with both a human-readable name for the event detected, along with the raw data behind it.

The following are supported human-readable names:
	Enter
	Esc

	Mouse Moved
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
	"strconv"
	"strings"
	"unicode"

	term "github.com/nsf/termbox-go"
)

const debugKeycodes = false

var directionNames = map[byte]Direction{
	65: Up,
	66: Down,
	67: Right,
	68: Left,
}

// Direction is Up, Down, Left, or Right
type Direction uint

// Arrow direction
const (
	Up Direction = iota
	Down
	Left
	Right
)

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(callback func(parsedData interface{}, rawData []byte)) {
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

		handle := func(parsedData interface{}) {
			callback(parsedData, data)
		}

		switch data[0] {
		case 13:
			handle(Enter{})
		case 195: // Alt
			parseAltLetter(data[1]-64, handle)
		case 27: // Escape code
			handleEscapeCode(data, handle)
		default:
			if len(data) == 1 {
				if data[0] <= 26 { // Ctrl
					letter := rune('A' + data[0] - 1)
					if letter == 'Q' { // exit upon Ctrl+Q
						return
					}
					handle(CtrlChar{letter})
				} else {
					letter := rune(data[0])
					handle(Character{letter})
				}
			}
		}

		if debugKeycodes {
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

func handleEscapeCode(data []byte, handle func(parsedData interface{})) {
	if len(data) == 1 { // Lone Esc Key
		handle("Esc")
		return
	}

	switch data[1] {
	case 79:
		direction := directionNames[data[2]]
		if len(data) == 15 { // scrolling
			switch data[2] {
			case 65:
				handle(ScrollUp{})
			case 66:
				handle(ScrollDown{})
			default:
				log.Printf("Unrecognized scroll code: %v", data)
			}
		} else { // arrow
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
				y, _ := strconv.Atoi(pieces[0])
				x, _ := strconv.Atoi(strings.TrimSuffix(pieces[1], "M"))
				handle(MouseDown{X: x, Y: y})
			case "35":
				y, _ := strconv.Atoi(pieces[0])
				x, _ := strconv.Atoi(strings.TrimSuffix(pieces[1], "M"))
				handle(MouseUp{X: x, Y: y})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		case 57: // Scrolling
			switch data[3] {
			case 54:
				handle(ScrollUp{})
			case 55:
				handle(ScrollDown{})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		default:
			arrow := directionNames[data[5]]
			switch data[4] {
			case 50:
				handle(ShiftArrow{arrow})
			case 51:
				handle(AltArrow{arrow})
			case 52:
				handle(AltShiftArrow{arrow})
			case 53:
				handle(CtrlArrow{arrow})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		}
	default:
		parseAltLetter(data[1], handle)
	}
}

func parseAltLetter(b byte, handle func(parsedData interface{})) {
	letter := rune(b)
	if unicode.IsUpper(letter) {
		handle(AltShiftChar{letter})
	} else {
		handle(AltChar{unicode.ToUpper(letter)})
	}
}
