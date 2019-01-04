/*
Package vterm provides a layer of abstraction between a channel of incoming text (possibly containing ANSI escape codes, et al) and a channel of outbound Char's.

A Char is a character printed using a given cursor (which is stored alongside the Char).
*/
package vterm

import (
	"fmt"
	"log"
	"sync"
	"time"
	"unicode"

	"github.com/aaronduino/i3-tmux/cursor"
)

// Char represents one character in the terminal's grid
type Char struct {
	Rune rune

	Cursor cursor.Cursor
}

/*
VTerm acts as a virtual terminal emulator between a shell and the host terminal emulator

It both transforms an inbound stream of bytes into Char's and provides the option of dumping all the Char's that need to be rendered to display the currently visible terminal window from scratch.
*/
type VTerm struct {
	w, h int

	buffer      [][]Char
	bufferMutux *sync.Mutex

	cursor cursor.Cursor

	in  <-chan rune
	out chan<- Char

	storedCursorX, storedCursorY int

	Selected bool
}

// NewVTerm returns a VTerm ready to be used by its exported methods
func NewVTerm(in <-chan rune, out chan<- Char) *VTerm {
	w := 20
	h := 20

	buffer := [][]Char{}
	for j := 0; j < h; j++ {
		row := []Char{}
		for i := 0; i < w; i++ {
			row = append(row, Char{
				Rune:   0,
				Cursor: cursor.Cursor{X: i, Y: j},
			})
		}
		buffer = append(buffer, row)
	}

	return &VTerm{
		w:           w,
		h:           h,
		buffer:      buffer,
		bufferMutux: &sync.Mutex{},
		cursor:      cursor.Cursor{X: 0, Y: 0},
		in:          in,
		out:         out,
		Selected:    false,
	}
}

// Reshape safely updates a VTerm's width & height
func (v *VTerm) Reshape(w, h int) {
	v.bufferMutux.Lock()
	v.w = w
	v.h = h
	v.bufferMutux.Unlock()
}

// RedrawWindow draws the entire visible window from scratch, sending the Char's to the scheduler via the out channel
func (v *VTerm) RedrawWindow() {
	v.bufferMutux.Lock()

	verticalArea := v.h
	if v.h > len(v.buffer) {
		verticalArea = len(v.buffer)
	}

	for _, row := range v.buffer[len(v.buffer)-verticalArea:] {
		for _, char := range row {
			// truncate characters past the width
			if char.Cursor.X > v.w {
				break
			}

			if char.Rune != 0 {
				v.out <- char
			}
		}
	}

	v.bufferMutux.Unlock()
}

// ProcessStream processes and transforms a process' stdout, turning it into a stream of Char's to be sent to the rendering scheduler
// This includes translating ANSI cursor coordinates and maintaining a scrolling buffer
func (v *VTerm) ProcessStream() {
	// blink cursor
	cursorTicker := time.NewTicker(time.Second / 2)
	cursorDone := make(chan bool)
	defer (func() {
		cursorDone <- true
	})()
	go (func() {
		visible := true
		for {
			select {
			case <-cursorTicker.C:
				v.bufferMutux.Lock()
				char := v.buffer[v.cursor.Y][v.cursor.X]
				if visible && v.Selected {
					char.Rune = '|'
					char.Cursor.Bg = cursor.Color{
						ColorMode: cursor.ColorBit3Normal,
						Code:      6,
					}
				} else {
					char.Rune = ' '
				}
				v.out <- char
				v.bufferMutux.Unlock()
				visible = !visible
				break
			case <-cursorDone:
				cursorTicker.Stop()
				return
			}
		}
	})()

	for {
		next, ok := <-v.in
		if !ok {
			return
		}

		switch next {
		case '\x1b':
			v.handleEscapeCode()
		case 8, 127:
			v.cursor.X--
			if v.cursor.X < 0 {
				v.cursor.X = 0
			}
		case '\n':
			v.cursor.Y++
			v.cursor.X = 0
		case '\r':
			v.cursor.X = 0
		default:
			if unicode.IsPrint(next) {
				v.bufferMutux.Lock()
				if v.cursor.X < 0 {
					v.cursor.X = 0
				}
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}

				char := Char{
					Rune:   next,
					Cursor: v.cursor,
				}

				if len(v.buffer)-1 < v.cursor.Y {
					for i := len(v.buffer); i < v.cursor.Y+1; i++ {
						v.buffer = append(v.buffer, []Char{Char{
							Rune:   0,
							Cursor: cursor.Cursor{X: 0, Y: i},
						}})
					}
				}
				if len(v.buffer[v.cursor.Y])-1 < v.cursor.X {
					for i := len(v.buffer[v.cursor.Y]); i < v.cursor.X+1; i++ {
						v.buffer[v.cursor.Y] = append(v.buffer[v.cursor.Y], Char{
							Rune:   0,
							Cursor: cursor.Cursor{X: i, Y: v.cursor.Y},
						})
					}
				}

				v.buffer[v.cursor.Y][v.cursor.X] = char
				v.out <- char
				v.cursor.X++
				v.bufferMutux.Unlock()
			} else {
				v.debug(fmt.Sprintf("%x    ", next))
			}
		}
	}
}

func (v *VTerm) handleEscapeCode() {
	next, ok := <-v.in
	if !ok {
		log.Fatal("not ok")
		return
	}

	switch next {
	case '[':
		v.handleCSISequence()
	default:
		v.debug("ESC Code: " + string(next))
	}
}

func (v *VTerm) handleCSISequence() {
	privateSequence := false

	parameterCode := ""
	for {
		next, ok := <-v.in
		if !ok {
			return
		}

		if unicode.IsDigit(next) || next == ';' || next == ' ' {
			parameterCode += string(next)
		} else if next == '?' {
			privateSequence = true
		} else if privateSequence {
			switch next {
			case 'h':
				switch parameterCode {
				case "25": // show cursor
					break // TODO
				case "1024": // enable alt screen buffer
					break // TODO
				case "2004": // disable alt screen buffer
					break // TODO
				}
			case 'l':
				switch parameterCode {
				case "25": // show cursor
					break // TODO
				case "1024": // enable alt screen buffer
					break // TODO
				case "2004": // disable alt screen buffer
					break // TODO
				}
			default:
				v.debug("CSI Private Code: " + parameterCode + string(next))
			}
			return
		} else {
			switch next {
			case 'A': // Cursor Up
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y -= seq[0]
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}
			case 'B': // Cursor Down
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y += seq[0]
			case 'C': // Cursor Right
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X += seq[0]
			case 'D': // Cursor Left
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X -= seq[0]
				if v.cursor.X < 0 {
					v.cursor.X = 0
				}
			case 'E': // Cursor Next Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y += seq[0]
				v.cursor.X = 0
			case 'F': // Cursor Previous Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y -= seq[0]
				v.cursor.X = 0
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}
			case 'G': // Cursor Horizontal Absolute
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X = seq[0] - 1
			case 'H', 'f': // Cursor Position
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y = seq[0] - 1
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}
				if len(seq) > 1 {
					v.cursor.X = seq[1] - 1
					if v.cursor.X < 0 {
						v.cursor.X = 0
					}
				}
			case 'J': // Erase in Display
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from cursor to end of screen
				case 1: // clear from cursor to beginning of screen
				case 2: // clear entire screen (and move cursor to top left?)
				case 3: // clear entire screen and delete all lines saved in scrollback buffer
				}
			case 'K': // Erase in Line
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from cursor to end of line
					for i := v.cursor.X; i < len(v.buffer[v.cursor.Y]); i++ {
						v.buffer[v.cursor.Y][i].Rune = 0
					}
				case 1: // clear from cursor to beginning of line
					for i := 0; i < v.cursor.X; i++ {
						v.buffer[v.cursor.Y][i].Rune = 0
					}
				case 2: // clear entire line; cursor position remains the same
					for i := 0; i < len(v.buffer[v.cursor.Y]); i++ {
						v.buffer[v.cursor.Y][i].Rune = 0
					}
				}
			case 'S': // Scroll Up; new lines added to bottom
			case 'T': // Scroll Down; new lines added to top
			case 'm': // Select Graphic Rendition
				v.handleSDR(parameterCode)
			case 's': // Save Cursor Position
				v.storedCursorX = v.cursor.X
				v.storedCursorY = v.cursor.Y
			case 'u': // Restore Cursor Positon
				v.cursor.X = v.storedCursorX
				v.cursor.Y = v.storedCursorY
			default:
				v.debug("CSI Code: " + string(next))
			}
			return
		}
	}
}

func (v *VTerm) handleSDR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	c := seq[0]

	switch c {
	case 0:
		v.cursor.Reset()
	case 1:
		v.cursor.Bold = true
	case 2:
		v.cursor.Faint = true
	case 3:
		v.cursor.Italic = true
	case 4:
		v.cursor.Underline = true
	case 5: // slow blink
	case 6: // rapid blink
	case 7: // swap foreground & background; see case 27
	case 8:
		v.cursor.Conceal = true
	case 9:
		v.cursor.CrossedOut = true
	case 10: // primary/default font
	case 22:
		v.cursor.Bold = false
		v.cursor.Faint = false
	case 23:
		v.cursor.Italic = false
	case 24:
		v.cursor.Underline = false
	case 25: // blink off
	case 27: // inverse off; see case 7
		break // TODO
	case 28:
		v.cursor.Conceal = false
	case 29:
		v.cursor.CrossedOut = false
	case 38: // set foreground color
	case 39: // default foreground color
	case 48: // set background color
	case 49: // default background color
	default:
		if c >= 30 && c <= 37 {
			if len(seq) > 1 && seq[1] == 1 {
				v.cursor.Fg = cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      int32(c - 30),
				}
			} else {
				v.cursor.Fg = cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      int32(c - 30),
				}
			}
		} else if c >= 40 && c <= 47 {
			if len(seq) > 1 && seq[1] == 1 {
				v.cursor.Bg = cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      int32(c - 40),
				}
			} else {
				v.cursor.Bg = cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      int32(c - 40),
				}
			}
		} else if c >= 90 && c <= 97 {
			v.cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit3Bright,
				Code:      int32(c - 90),
			}
		} else if c >= 100 && c <= 107 {
			v.cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit3Bright,
				Code:      int32(c - 100),
			}
		} else {
			v.debug("SGR Code: " + string(parameterCode))
		}
	}
}
