package keypress

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	gc "github.com/rthornton128/goncurses"
)

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

	// if should {
	// 	fmt.Print("\033[?1006h")
	// } else {
	// 	fmt.Print("\033[?1006l")
	// }

	isProcessingMouse = should
}

// this is a regularly reset buffer of what we've collected so far
var data []byte

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

func pullByte() byte {
	b := byte(stdscr.GetChar())
	data = append(data, b)
	return b
}

var stdscr *gc.Window

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(s *gc.Window, callback func(parsedData interface{}, rawData []byte)) {
	stdscr = s
	// go func() {
	// 	for {
	// 		c := make(chan os.Signal, 1)
	// 		signal.Notify(c, syscall.SIGWINCH)
	// 		<-c
	// 		w, h, _ := GetTermSize()
	// 		callback(Resize{W: w, H: h}, []byte{})
	// 	}
	// }()

	for {
		stdscr.Timeout(-1)
		data = []byte{}
		key := pullByte()

		stdscr.Timeout(5)

		handle := func(x interface{}) {
			callback(x, data)
		}

		switch key {
		case 0:
		case 13:
			handle(Enter{})
		case 195: // Alt
			parseAltLetter(data[1]-64, handle)
		case 27: // Escape code
			handleEscapeCode(handle, callback)
		default:
			if key <= 26 { // Ctrl
				letter := rune('A' + key - 1)
				if letter == 'Q' { // exit upon Ctrl+Q
					return
				}
				handle(CtrlChar{letter})
			} else {
				handle(Character{Char: rune(key)})
			}
		}
	}
}

func handleEscapeCode(handle func(parsedData interface{}), callback func(parsedData interface{}, rawData []byte)) {
	b := pullByte()
	switch b {
	case 0:
		handle(Esc{})
	case 13:
		log.Println("AltChar{'\\n'}")
		handle(AltChar{'\n'})
	case 27:
		switch pullByte() {
		case 91:
			switch pullByte() {
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
		default:
			log.Println("Unhandled double escape:", data)
		}
	case 79:
		log.Println("79!?!?")
		// direction := directionNames[data[2]]
		// if len(data) == 15 { // scrolling
		// 	switch data[2] {
		// 	case 65:
		// 		handle(ScrollUp{})
		// 	case 66:
		// 		handle(ScrollDown{})
		// 	default:
		// 		log.Printf("Unrecognized scroll code: %v", data)
		// 	}
		// } else { // arrow
		// 	handle(direction)
		// }
	case 91: // '['
		switch pullByte() {
		case 49:
			switch pullByte() {
			case 59:
				switch pullByte() {
				case 49:
					switch pullByte() {
					case 48:
						handle(AltShiftArrow{Direction: directionNames[pullByte()]})
					default:
						log.Println("unhandled: ", data)
					}
				case 51:
					handle(AltArrow{Direction: directionNames[pullByte()]})
				case 52:
					handle(AltShiftArrow{Direction: directionNames[pullByte()]})
				default:
					log.Println("unhandled: ", data)
				}
			default:
				log.Println("unhandled: ", data)
			}
		case 51: // Mouse??
			switch pullByte() {
			case 126:
				handle(Character{Char: 127})
			default:
				log.Println("unhandled: ", data)
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
			switch pullByte() {
			case 54:
				switch pullByte() {
				case 52:
					handle(ScrollUp{})
				case 53:
					handle(ScrollDown{})
				default:
					log.Printf("Unrecognized scroll code': %v", data)
				}
			default:
				log.Println("unhandled: ", data)
			}
		case 65:
			callback(Arrow{Direction: Up}, []byte{27, 79, 65})
		case 66:
			callback(Arrow{Direction: Down}, []byte{27, 79, 66})
		case 67:
			callback(Arrow{Direction: Right}, []byte{27, 79, 67})
		case 68:
			callback(Arrow{Direction: Left}, []byte{27, 79, 68})
		case 77:
			switch pullByte() {
			// case 32:
			// 	handle(MouseDown{X: int(data[4] - 32), Y: int(data[5] - 32)})
			// case 35:
			// 	handle(MouseUp{X: int(data[4] - 32), Y: int(data[5] - 32)})
			case 96:
				handle(ScrollUp{})
			case 97:
				handle(ScrollDown{})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		default:
			switch pullByte() {
			case 50:
				handle(ShiftArrow{directionNames[pullByte()]})
			case 51:
				handle(AltArrow{directionNames[pullByte()]})
			case 52:
				handle(AltShiftArrow{directionNames[pullByte()]})
			case 53:
				handle(CtrlArrow{directionNames[pullByte()]})
			default:
				log.Printf("Unrecognized keycode: %v", data)
			}
		}
	case 102:
		handle(AltChar{Char: 'F'})
	default:
		parseAltLetter(pullByte(), handle)
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
