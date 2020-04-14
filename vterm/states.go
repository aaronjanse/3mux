package vterm

import (
	"fmt"
	"log"
	"unicode"
	"strings"

	"github.com/aaronjanse/3mux/render"
)

type Parser struct {
	state State

	private *rune
	intermediate string
	params string
	final *rune
}

type State int

const (
	StateGround = iota
	StateEscape
	StateCsiEntry
	StateCsiParam
	StateOscString
)

func (v *VTerm) ProcessStream() {
	for {
		r, ok := v.pullRune()
		// log.Printf("# %s (%d)", string(r), r)
		if !ok {
			return
		}
		v.Anywhere(r)
	}
}

func (v *VTerm) Anywhere(r rune) {
	switch r {
	case 0x00:
	case 0x1B:
		v.DoClear()
		v.parser.state = StateEscape
	case 0x8D: // Reverse Index
		if v.Cursor.Y == 0 {
			v.scrollDown(1)
		} else {
			v.shiftCursorY(-1)
		}
		v.RedrawWindow()
		v.parser.state = StateGround
	case 0x9B:
		v.DoClear()
		v.parser.state = StateCsiEntry
	case 0x9C:
		v.parser.state = StateGround
	case 0x9D:
		v.parser.state = StateOscString
	default:
		switch v.parser.state {
		case StateGround:
			v.StateGround(r)
		case StateEscape:
			v.StateEscape(r)
		case StateCsiEntry:
			v.StateCsiEntry(r)
		case StateCsiParam:
			v.StateCsiParam(r)
		case StateOscString:
			v.StateOscString(r)
		default:
			log.Printf("? STATE %d", v.parser.state)
		}
	}
}

func (v *VTerm) StateGround(r rune) {
	switch {
	case 8 == r:
		if v.Cursor.X > 0 {
			v.shiftCursorX(-1)
		}
	case '\n' == r:
		if v.Cursor.Y == v.scrollingRegion.bottom {
			v.scrollUp(1)
		} else {
			v.shiftCursorY(1)
		}
	case '\r' == r:
		v.setCursorX(0)
	case '\t' == r:
		tabWidth := 8 // FIXME
		v.Cursor.X += tabWidth - (v.Cursor.X % tabWidth)
	case unicode.IsPrint(r):
		if v.Cursor.X > v.w-1 {
			v.setCursorX(0)
			if v.Cursor.Y == v.scrollingRegion.bottom {
				v.scrollUp(1)
			} else {
				v.shiftCursorY(1)
			}
		}
		v.putChar(r)
	default:
		log.Printf("? GROUND %s (%d)", string(r), r)
	}
}

func (v *VTerm) StateEscape(r rune) {
	switch {
	case strings.Contains("DEHMNOPVWXZ[\\]^_", string(r)):
		v.Anywhere(r + 0x40)
	case 0x30 <= r && r <= 0x4F || 0x51 <= r && r <= 0x57:
		// TODO: v.DispatchEsc()
		v.parser.state = StateGround
	default:
		log.Printf("? ESC %s", string(r))
	}
}

func (v *VTerm) StateCsiEntry(r rune) {
	switch {
	case 0x30 <= r && r <= 0x39 || r == 0x3B:
		v.parser.params += string(r)
		v.parser.state = StateCsiParam
	case 0x3C <= r && r <= 0x3F:
		v.parser.intermediate += string(r)
		v.parser.state = StateCsiParam
	case 0x40 <= r && r <= 0x7E:
		v.parser.final = &r
		v.DispatchCsi()
		v.parser.state = StateGround
	}
}

func (v *VTerm) StateCsiParam(r rune) {
	switch {
	case 0x30 <= r && r <= 0x39 || r == 0x3b:
		v.parser.params += string(r)
	case 0x40 <= r && r <= 0x7E:
		v.parser.final = &r
		v.DispatchCsi()
		v.parser.state = StateGround
	}
}

func (v *VTerm) StateOscString(r rune) {
	switch {
	case 0x07 == r: // FIXME: this is weird
		v.parser.state = StateGround
	}
}

func (v *VTerm) DoClear() {
	v.parser.private = nil
	v.parser.intermediate = ""
	v.parser.params = ""
	v.parser.final = nil
}

func (v *VTerm) DispatchCsi() {
	switch v.parser.intermediate {
	case "?":
		switch *v.parser.final {
			case 'h': // DECSET
				switch v.parser.params {
				// case "1": // application arrow keys (DECCKM)
				// case "7": // Auto-wrap Mode (DECAWM)
				// case "12": // start blinking Cursor
				// case "25": // show Cursor

				// FIXME: distinguish between these codes
				case "1049", "1047", "47": // switch to alt screen buffer
					if !v.usingAltScreen {
						v.screenBackup = v.Screen
					}
				// case "2004": // enable bracketed paste mode
				default:
					log.Printf("? DECSET %s", v.parser.params)
				}
			case 'l': // generally disables features
				switch v.parser.params { // DECRST
				// case "1": // Normal Cursor keys (DECCKM)
				// case "7": // No Auto-wrap Mode (DECAWM)
				// case "12": // stop blinking Cursor
				// case "25": // hide Cursor

				// FIXME: distinguish between these codes
				case "1049", "1047", "47": // switch to normal screen buffer
					if v.usingAltScreen {
						v.Screen = v.screenBackup
					}
				// case "2004": // disable bracketed paste mode
				default:
					log.Printf("? DECRST %s", v.parser.params)
				}
			default:
				log.Printf("? CSI ? %s %s", v.parser.params, string(*v.parser.final))
			}
	case "":
		switch *v.parser.final {
			case 'A': // Cursor Up
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorY(-n)
				}
			case 'B': // Cursor Down
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorY(n)
				}
			case 'C': // Cursor Right
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorX(n)
				}
			case 'D': // Cursor Left
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorX(-n)
				}
			case 'd': // Vertical Line Position Absolute (VPA)
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.setCursorY(seq[0] - 1)
			case 'E': // Cursor Next Line
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.shiftCursorY(seq[0])
				v.setCursorX(0)
			case 'F': // Cursor Previous Line
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.setCursorX(0)
				v.shiftCursorY(-seq[0])
			case 'G': // Cursor Horizontal Absolute
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.setCursorX(seq[0] - 1)
			case 'H', 'f': // Cursor Position
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				if v.parser.params == "" {
					v.setCursorPos(0, 0)
				} else {
					v.setCursorY(seq[0] - 1)
					if len(seq) > 1 {
						v.setCursorX(seq[1] - 1)
					}
				}
			case 'J': // Erase in Display
				v.handleEraseInDisplay(v.parser.params)
			case 'K': // Erase in Line
				v.handleEraseInLine(v.parser.params)
			case 'M': // Delete Lines; https://vt100.net/docs/vt510-rm/DL.html
				n := parseSemicolonNumSeq(v.parser.params, 1)[0]
				log.Printf("DL %s (%d)", v.parser.params, n)

				newLines := make([][]render.Char, n)
				for i := range newLines {
					newLines[i] = make([]render.Char, v.w)
				}

				v.Screen = append(append(append(
					v.Screen[:v.Cursor.Y],
					v.Screen[v.Cursor.Y+n:v.scrollingRegion.bottom+1]...),
					newLines...),
					v.Screen[v.scrollingRegion.bottom+1:]...)

				if !v.usingSlowRefresh {
					v.RedrawWindow()
				}
			case 'n': // Device Status Report
				seq := parseSemicolonNumSeq(v.parser.params, 0)
				switch seq[0] {
				case 6:
					response := fmt.Sprintf("\x1b[%d;%dR", v.Cursor.Y+1, v.Cursor.X+1)
					for _, r := range response {
						v.out <- r
					}
				default:
					log.Println("Unrecognized DSR code", seq)
				}
			case 'r': // Set Scrolling Region
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.scrollingRegion.top = seq[0] - 1
				if len(seq) > 1 {
					v.scrollingRegion.bottom = seq[1] - 1
				} else {
					v.scrollingRegion.bottom = v.h + 1
				}
				v.setCursorPos(0, 0)
			case 'S': // Scroll Up; new lines added to bottom
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.scrollUp(seq[0])
			case 'T': // Scroll Down; new lines added to top
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.scrollDown(seq[0])
			// case 't': // Window Manipulation
			// 	// TODO
			case 'L': // Insert Lines; https://vt100.net/docs/vt510-rm/IL.html
				seq := parseSemicolonNumSeq(v.parser.params, 1)
				v.setCursorX(0)

				n := seq[0]
				newLines := make([][]render.Char, n)
				for i := range newLines {
					newLines[i] = make([]render.Char, v.w)
				}

				newLines = append(append(
					newLines,
					v.Screen[v.Cursor.Y:v.scrollingRegion.bottom-n+1]...),
					v.Screen[v.scrollingRegion.bottom+1:]...)

				copy(v.Screen[v.Cursor.Y:], newLines)

				v.RedrawWindow()
			case 'm': // Select Graphic Rendition
				v.handleSGR(v.parser.params)
			case 's': // Save Cursor Position
				v.storedCursorX = v.Cursor.X
				v.storedCursorY = v.Cursor.Y
			case 'u': // Restore Cursor Positon
				v.setCursorPos(v.storedCursorX, v.storedCursorY)
			default:
				log.Printf("? CSI %s %s", v.parser.params, string(*v.parser.final))
			}
	}
}
