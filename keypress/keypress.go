package keypress

import (
	"unicode"

	term "github.com/gdamore/tcell"
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

var Scr term.Screen

// ShouldProcessMouse controls whether typical mouse actions (e.g. selection) should be disabled in exhange for i3-mux-specific features
func ShouldProcessMouse(should bool) {
	if isProcessingMouse == should {
		return
	}

	if should {
		Scr.EnableMouse()
	} else {
		Scr.DisableMouse()
	}

	isProcessingMouse = should
}

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(callback func(parsedData interface{}, rawData []byte), scr term.Screen) {
	Scr = scr

	ShouldProcessMouse(true)

	// show cursor, make it blink
	// fmt.Print("\033[?25h\033[?12h") // EXPLAIN: do we need this?

	for {
		// raw := make([]byte, 16) // EXPLAIN: why 16?
		ev := Scr.PollEvent()
		switch k := ev.(type) {
		case *term.EventKey:
			switch k.Key() {
			case term.KeyUp:
				switch k.Modifiers() {
				case term.ModAlt:
					callback(AltArrow{Direction: Up}, []byte{})
				case term.ModAlt | term.ModShift:
					callback(AltShiftArrow{Direction: Up}, []byte{})
				}
			case term.KeyDown:
				switch k.Modifiers() {
				case term.ModAlt:
					callback(AltArrow{Direction: Down}, []byte{})
				case term.ModAlt | term.ModShift:
					callback(AltShiftArrow{Direction: Down}, []byte{})
				}
			case term.KeyLeft:
				switch k.Modifiers() {
				case term.ModAlt:
					callback(AltArrow{Direction: Left}, []byte{})
				case term.ModAlt | term.ModShift:
					callback(AltShiftArrow{Direction: Left}, []byte{})
				}
			case term.KeyRight:
				switch k.Modifiers() {
				case term.ModAlt:
					callback(AltArrow{Direction: Right}, []byte{})
				case term.ModAlt | term.ModShift:
					callback(AltShiftArrow{Direction: Right}, []byte{})
				}
			default:
				r := k.Rune()
				switch k.Modifiers() {
				case term.ModAlt:
					if unicode.IsUpper(r) {
						callback(AltShiftChar{Char: k.Rune()}, []byte{byte(r)})
					} else {
						callback(AltChar{Char: unicode.ToUpper(r)}, []byte{byte(r)})
					}
				default:
					callback(Character{Char: k.Rune()}, []byte{byte(r)})
				}
			}
		case *term.EventMouse:
			x, y := k.Position()
			switch k.Buttons() {
			case term.ButtonNone:
				callback(MouseUp{X: x, Y: y}, []byte{})
			case term.Button1:
				callback(MouseDown{X: x, Y: y}, []byte{})
			}
		}
	}
}
