package vterm

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/aaronduino/i3-tmux/cursor"
)

// ProcessStream processes and transforms a process' stdout, turning it into a stream of Char's to be sent to the rendering scheduler
// This includes translating ANSI cursor coordinates and maintaining a scrolling buffer
func (v *VTerm) ProcessStream() {
	for {
		next, ok := <-v.in
		if !ok {
			return
		}

		// <-time.NewTimer(time.Second / 32).C

		if next > 127 {
			value := []byte{byte(next)}

			leadingHex := next >> 4
			// v.debug(fmt.Sprintf("%x", next>>4))
			switch leadingHex {
			case 12: // 1100
				value = append(value, byte(<-v.in))
			case 14: // 1110
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
			case 15: // 1111
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
			}

			// v.debug(fmt.Sprintf("%x", value))

			next, _ = utf8.DecodeRune(value)
		}

		switch next {
		case '\x1b':
			v.handleEscapeCode()
		case 8:
			v.cursor.X--
			if v.cursor.X < 0 {
				v.cursor.X = 0
			}
			v.updateCursor()
		case '\n':
			// v.cursor.X = 0
			if v.cursor.Y == v.scrollingRegion.bottom {
				// disable scrollback if using alt screen
				if !v.usingAltScreen {
					// v.scrollback = append(v.scrollback, v.screen[len(v.screen)-1:]...)
					v.scrollback = append(v.scrollback, v.screen[v.scrollingRegion.top])
				}

				// v.screen = append(v.screen[:len(v.screen)-1], []Char{})
				v.screen = append(append(append(
					v.screen[:v.scrollingRegion.top],
					v.screen[v.scrollingRegion.top+1:v.scrollingRegion.bottom+1]...),
					[]Char{}),
					v.screen[v.scrollingRegion.bottom+1:]...)

				v.RedrawWindow()
			} else {
				v.cursor.Y++
			}
			v.updateCursor()
		case '\r':
			v.cursor.X = 0
			v.updateCursor()
		default:
			if unicode.IsPrint(next) {
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

				if len(v.screen)-1 < v.cursor.Y {
					for i := len(v.screen); i < v.cursor.Y+1; i++ {
						v.screen = append(v.screen, []Char{Char{
							Rune: 0,
						}})
					}
				}
				if len(v.screen[v.cursor.Y])-1 < v.cursor.X {
					for i := len(v.screen[v.cursor.Y]); i < v.cursor.X+1; i++ {
						v.screen[v.cursor.Y] = append(v.screen[v.cursor.Y], Char{
							Rune: 0,
						})
					}
				}

				v.screen[v.cursor.Y][v.cursor.X] = char

				if v.cursor.X < v.w && v.cursor.Y < v.h {
					char.Cursor.X = v.cursor.X
					char.Cursor.Y = v.cursor.Y
					v.out <- char
				}

				v.cursor.X++
				v.updateCursor()
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
	case '(': // Character set
		<-v.in
		// TODO
	default:
		v.debug("ESC Code: " + string(next))
	}
}

func (v *VTerm) handleCSISequence() {
	privateSequence := false

	// <-time.NewTimer(time.Second / 2).C

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
				case "1": // application arrow keys (DECCKM)
				case "7": // Auto-wrap Mode (DECAWM)
				case "12": // start blinking cursor
				case "25": // show cursor
					// v.StartBlinker()
				case "1049", "1047", "47": // switch to alt screen buffer
					if !v.usingAltScreen {
						v.screenBackup = v.screen
					}
				case "2004": // enable bracketed paste mode
				default:
					v.debug("CSI Private H Code: " + parameterCode + string(next))
				}
			case 'l':
				switch parameterCode {
				case "1": // Normal cursor keys (DECCKM)
				case "7": // No Auto-wrap Mode (DECAWM)
				case "12": // stop blinking cursor
				case "25": // hide cursor
					// v.StopBlinker()
				case "1049", "1047", "47": // switch to normal screen buffer
					if v.usingAltScreen {
						v.screen = v.screenBackup
					}
				case "2004": // disable bracketed paste mode
					// TODO
				default:
					v.debug("CSI Private L Code: " + parameterCode + string(next))
				}
			default:
				v.debug("CSI Private Code: " + parameterCode + string(next))
			}
			return
		} else {
			// if next != 'H' && next != 'C' && next != 'G' && next != 'm' {
			// 	v.debug(string(next))
			// }
			// v.debug(string(next))
			switch next {
			case 'A': // Cursor Up
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y -= seq[0]
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}
				v.updateCursor()
			case 'B': // Cursor Down
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y += seq[0]
				v.updateCursor()
			case 'C': // Cursor Right
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X += seq[0]
				v.updateCursor()
			case 'D': // Cursor Left
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X -= seq[0]
				if v.cursor.X < 0 {
					v.cursor.X = 0
				}
				v.updateCursor()
			case 'd': // Vertical Line Position Absolute (VPA)
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y = seq[0] - 1
				v.updateCursor()
			case 'E': // Cursor Next Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y += seq[0]
				v.cursor.X = 0
				v.updateCursor()
			case 'F': // Cursor Previous Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.Y -= seq[0]
				v.cursor.X = 0
				if v.cursor.Y < 0 {
					v.cursor.Y = 0
				}
				v.updateCursor()
			case 'G': // Cursor Horizontal Absolute
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X = seq[0] - 1
			case 'H', 'f': // Cursor Position
				seq := parseSemicolonNumSeq(parameterCode, 1)
				if parameterCode == "" {
					v.cursor.X = 0
					v.cursor.Y = 0
				} else {
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
				}
				v.updateCursor()
			case 'J': // Erase in Display
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from cursor to end of screen
					for i := v.cursor.X; i < len(v.screen[v.cursor.Y]); i++ {
						v.screen[v.cursor.Y][i].Rune = 0
					}
					if v.cursor.Y+1 < len(v.screen) {
						for j := v.cursor.Y; j < len(v.screen); j++ {
							for i := 0; i < len(v.screen[j]); i++ {
								v.screen[j][i].Rune = 0
							}
						}
					}
				case 1: // clear from cursor to beginning of screen
					for j := 0; j < v.cursor.Y; j++ {
						for i := 0; i < len(v.screen[j]); j++ {
							v.screen[j][i].Rune = 0
						}
					}
					for i := 0; i < v.cursor.X; i++ {
						v.screen[v.cursor.Y][i].Rune = 0
					}
				case 2: // clear entire screen (and move cursor to top left?)
					for i := range v.screen {
						for j := range v.screen[i] {
							v.screen[i][j].Rune = ' '
						}
					}
					v.cursor.X = 0
					v.cursor.Y = 0
				case 3: // clear entire screen and delete all lines saved in scrollback buffer
					// TODO
				}
				v.RedrawWindow()
			case 'K': // Erase in Line
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from cursor to end of line
					for i := v.cursor.X; i < len(v.screen[v.cursor.Y]); i++ { // FIXME: sometimes crashes
						v.screen[v.cursor.Y][i].Rune = 0
					}
				case 1: // clear from cursor to beginning of line
					for i := 0; i < v.cursor.X; i++ {
						v.screen[v.cursor.Y][i].Rune = 0
					}
				case 2: // clear entire line; cursor position remains the same
					for i := 0; i < len(v.screen[v.cursor.Y]); i++ {
						v.screen[v.cursor.Y][i].Rune = 0
					}
				}
				v.RedrawWindow()
			case 'r': // Set Scrolling Region
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.scrollingRegion.top = seq[0] - 1
				if len(seq) > 1 {
					v.scrollingRegion.bottom = seq[1] - 1
				} else {
					v.scrollingRegion.bottom = v.h - 1
				}
				v.cursor.X = 0
				v.cursor.Y = 0
			case 'S': // Scroll Up; new lines added to bottom
				seq := parseSemicolonNumSeq(parameterCode, 1)
				numLines := seq[0]
				// v.scrollback = append(v.scrollback, v.screen[:numLines]...)
				// v.screen = v.screen[numLines:]
				if !v.usingAltScreen {
					// v.scrollback = append(v.scrollback, v.screen[len(v.screen)-1:]...)
					v.scrollback = append(v.scrollback, v.screen[v.scrollingRegion.top:v.scrollingRegion.top+numLines]...)
				}

				newLines := make([][]Char, numLines)

				// v.screen = append(v.screen[:len(v.screen)-1], []Char{})
				v.screen = append(append(append(
					v.screen[:v.scrollingRegion.top],
					v.screen[v.scrollingRegion.top+numLines:v.scrollingRegion.bottom+1]...),
					newLines...),
					v.screen[v.scrollingRegion.bottom+1:]...)

				v.RedrawWindow()
			case 'T': // Scroll Down; new lines added to top
				seq := parseSemicolonNumSeq(parameterCode, 1)
				numLines := seq[0]
				// v.screen = append(v.scrollback[len(v.scrollback)-numLines:], v.screen...)
				// v.scrollback = v.scrollback[:len(v.scrollback)-numLines]

				newLines := make([][]Char, numLines)

				// v.screen = append(v.screen[:len(v.screen)-1], []Char{})
				v.screen = append(append(append(
					v.screen[:v.scrollingRegion.top],
					newLines...),
					v.screen[v.scrollingRegion.top:v.scrollingRegion.bottom-numLines]...),
					v.screen[v.scrollingRegion.bottom+1:]...)

				v.RedrawWindow()
			case 'L': // Insert Lines
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.cursor.X = 0

				if v.cursor.Y < v.scrollingRegion.top || v.cursor.Y > v.scrollingRegion.bottom {
					v.debug("unhandled insert line operation")
					return
				}

				numLines := seq[0]
				newLines := make([][]Char, numLines)

				above := [][]Char{}
				if v.cursor.Y > 0 {
					above = v.screen[:v.cursor.Y]
				}

				v.screen = append(append(append(
					above,
					newLines...),
					v.screen[v.cursor.Y:v.scrollingRegion.bottom-numLines+1]...),
					v.screen[v.scrollingRegion.bottom+1:]...)

				v.cursor.Y++
				v.updateCursor()
				v.RedrawWindow()
				// v.debug("                        " + strconv.Itoa(v.cursor.Y))
			case 'm': // Select Graphic Rendition
				v.handleSDR(parameterCode)
			case 's': // Save Cursor Position
				v.storedCursorX = v.cursor.X
				v.storedCursorY = v.cursor.Y
				v.updateCursor()
			case 'u': // Restore Cursor Positon
				v.cursor.X = v.storedCursorX
				v.cursor.Y = v.storedCursorY
				v.updateCursor()
			default:
				v.debug("CSI Code: " + string(next) + " ; " + parameterCode)
			}
			return
		}
	}
}

func (v *VTerm) handleSDR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	if parameterCode == "39;49" {
		v.cursor.Fg.ColorMode = cursor.ColorNone
		v.cursor.Bg.ColorMode = cursor.ColorNone
		return
	}

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
		// TODO
	case 28:
		v.cursor.Conceal = false
	case 29:
		v.cursor.CrossedOut = false
	case 38: // set foreground color
		if seq[1] == 5 {
			v.cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 39: // default foreground color
		v.cursor.Fg.ColorMode = cursor.ColorNone
	case 48: // set background color
		if seq[1] == 5 {
			v.cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 49: // default background color
		v.cursor.Bg.ColorMode = cursor.ColorNone
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
