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
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	term "github.com/nsf/termbox-go"
	"golang.org/x/crypto/ssh/terminal"
)

const debugKeycodes = false

var isProcessingMouse = true

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

// ShouldProcessMouse controls whether typical mouse actions (e.g. selection) should be disabled in exhange for i3-mux-specific features
func ShouldProcessMouse(should bool) {
	if isProcessingMouse == should {
		return
	}

	if should {
		term.SetInputMode(term.InputEsc | term.InputAlt | term.InputMouse)
	} else {
		term.SetInputMode(term.InputEsc | term.InputAlt)
	}

	isProcessingMouse = should
}

// this is a regularly reset buffer of what we've collected so far
var data []byte

func pullByte() byte {
	var b = make([]byte, 1)
	os.Stdin.Read(b)
	data = append(data, b[0])
	return b[0]
}

var oldState *terminal.State

// Shutdown cleans up the terminal state
func Shutdown() {
	ShouldProcessMouse(false)
	term.Close()
}

// GetTermSize returns the terminal dimensions w, h, err
func GetTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(callback func(parsedData interface{}, rawData []byte)) {
	err := term.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	ShouldProcessMouse(true)

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			w, h, _ := GetTermSize()
			callback(Resize{W: w, H: h}, []byte{})
		}
	}()

	ShouldProcessMouse(false)

	// show cursor, make it blink
	fmt.Print("\033[?25h\033[?12h") // EXPLAIN: do we need this?

	for {
		raw := make([]byte, 16) // EXPLAIN: why 16?
		ev := term.PollRawEvent(raw)

		data := raw[:ev.N]

		handle := func(parsedData interface{}) {
			callback(parsedData, data)
		}

		if ev.N == 0 && ev.Type == term.EventResize {
			handle(Resize{W: ev.Width, H: ev.Height})
		} else {
			switch data[0] {
			case 13:
				handle(Enter{})
			case 195: // Alt
				parseAltLetter(data[1]-64, handle)
			case 27: // Escape code
				handleEscapeCode(data, handle)
			default:
				// log.Println("uncaught:", data)
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
				} else {
					for _, b := range data {
						if b == 0 {
							break
						}
						callback(Character{rune(b)}, []byte{b})
					}
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
	case 13:
		handle(AltChar{'\n'})
	case 27:
		if len(data) == 4 && data[2] == 79 {
			switch data[3] {
			case 65:
				log.Println("Terminal.app!")
			case 66:
				log.Println("Terminal.app!")
			default:
				log.Println("Unhandled ?arrow?:", data)
			}
		} else if len(data) == 4 && data[2] == 91 {
			switch data[3] {
			case 65:
				handle(AltArrow{Direction: Up})
			case 66:
				handle(AltArrow{Direction: Down})
			case 67:
				handle(AltArrow{Direction: Right})
			case 68:
				handle(AltArrow{Direction: Left})
			default:
				log.Println("Unhandled ?arrow?:", data)
			}
		} else {
			log.Println("Unhandled double escape:", data)
		}
	case 79:
		direction := directionNames[data[2]]
		if len(data) > 3 { // scrolling
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
	case 91: // '['
		switch data[2] {
		case 49:
			if len(data) == 7 && data[3] == 59 && data[4] == 49 && data[5] == 48 {
				switch data[6] {
				case 65:
					handle(AltShiftArrow{Direction: Up})
				case 66:
					handle(AltShiftArrow{Direction: Down})
				case 67:
					handle(AltShiftArrow{Direction: Right})
				case 68:
					handle(AltShiftArrow{Direction: Left})
				default:
					log.Println("Unhandled alt shift arrow:", data)
				}
			} else if len(data) == 6 && data[3] == 59 && data[4] == 51 {
				switch data[5] {
				case 65:
					handle(AltArrow{Direction: Up})
				case 66:
					handle(AltArrow{Direction: Down})
				case 67:
					handle(AltArrow{Direction: Right})
				case 68:
					handle(AltArrow{Direction: Left})
				default:
					log.Println("Unhandled alt arrow:", data)
				}
			} else if len(data) == 6 && data[3] == 59 && data[4] == 52 {
				switch data[5] {
				case 65:
					handle(AltShiftArrow{Direction: Up})
				case 66:
					handle(AltShiftArrow{Direction: Down})
				case 67:
					handle(AltShiftArrow{Direction: Right})
				case 68:
					handle(AltShiftArrow{Direction: Left})
				default:
					log.Println("Unhandled alt arrow:", data)
				}
			} else {
				log.Println("Unhandled almost-shift-arrow:", data)
			}
		case 51: // Mouse
			if data[3] == 126 { // Delete
				handle(Character{Char: 127})
			} else {
				code := string(data[2:])
				code = strings.TrimSuffix(code, "M") // NOTE: are there other codes we are forgetting about?
				pieces := strings.Split(code, ";")
				switch pieces[0] {
				case "32":
					x, _ := strconv.Atoi(pieces[1])
					y, _ := strconv.Atoi(strings.TrimSuffix(pieces[2], "M"))
					handle(MouseDown{X: x, Y: y})
				case "35":
					x, _ := strconv.Atoi(pieces[1])
					y, _ := strconv.Atoi(strings.TrimSuffix(pieces[2], "M"))
					handle(MouseUp{X: x, Y: y})
				default:
					log.Printf("Unrecognized keycode: %v", data)
				}
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
		case 60:
			switch data[3] {
			case 54:
				switch data[4] {
				case 52:
					handle(ScrollUp{})
				case 53:
					handle(ScrollDown{})
				default:
					log.Printf("Unrecognized scroll code': %v", data)
				}
			case 48:
				code := string(data[5:])
				mode := data[len(data)-1]
				pieces := strings.Split(code[:len(code)-1], ";")
				var x int
				if len(pieces[0]) == 0 {
					x = 1
				} else {
					x, _ = strconv.Atoi(pieces[0])
				}
				y, _ := strconv.Atoi(pieces[1])
				switch mode {
				case 'M':
					handle(MouseDown{X: x, Y: y})
				case 'm':
					handle(MouseUp{X: x, Y: y})
				default:
					log.Printf("Unrecognzied mouse code: %v", data)
				}
			}
		case 65:
			handle(Arrow{Direction: Up})
		case 66:
			handle(Arrow{Direction: Down})
		case 67:
			handle(Arrow{Direction: Right})
		case 68:
			handle(Arrow{Direction: Left})
		case 77:
			switch data[3] {
			case 32:
				handle(MouseDown{X: int(data[4] - 32), Y: int(data[5] - 32)})
			case 35:
				handle(MouseUp{X: int(data[4] - 32), Y: int(data[5] - 32)})
			case 96:
				handle(ScrollUp{})
			case 97:
				handle(ScrollDown{})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		default:
			if len(data) < 6 {
				log.Printf("Unrecognized arrow %v", data)
				return
			}
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
	case 102:
		handle(AltChar{Char: 'F'})
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
