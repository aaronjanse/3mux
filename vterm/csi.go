package vterm

import (
	"unicode"

	"github.com/aaronduino/i3-tmux/render"
)

func (v *VTerm) handleCSISequence() {
	privateSequence := false

	// <-time.NewTimer(time.Millisecond * 10).C

	parameterCode := ""
	for {
		next, ok := v.pullByte()
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
				case "12": // start blinking Cursor
				case "25": // show Cursor
				case "1049", "1047", "47": // switch to alt screen buffer
					if !v.usingAltScreen {
						v.screenBackup = v.screen
					}
				case "2004": // enable bracketed paste mode
				default:
					// v.debug("CSI Private H Code: " + parameterCode + string(next))
				}
			case 'l':
				switch parameterCode {
				case "1": // Normal Cursor keys (DECCKM)
				case "7": // No Auto-wrap Mode (DECAWM)
				case "12": // stop blinking Cursor
				case "25": // hide Cursor
				case "1049", "1047", "47": // switch to normal screen buffer
					if v.usingAltScreen {
						v.screen = v.screenBackup
					}
				case "2004": // disable bracketed paste mode
				default:
					// v.debug("CSI Private L Code: " + parameterCode + string(next))
				}
			default:
				// v.debug("CSI Private Code: " + parameterCode + string(next))
			}
			return
		} else {
			switch next {
			case 'A': // Cursor Up
				seq := parseSemicolonNumSeq(parameterCode, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorY(-n)
				}
			case 'B': // Cursor Down
				seq := parseSemicolonNumSeq(parameterCode, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorY(n)
				}
			case 'C': // Cursor Right
				seq := parseSemicolonNumSeq(parameterCode, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorX(n)
				}
			case 'D': // Cursor Left
				seq := parseSemicolonNumSeq(parameterCode, 1)
				n := seq[0]
				if n > 0 {
					v.shiftCursorX(-n)
				}
			case 'd': // Vertical Line Position Absolute (VPA)
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.setCursorY(seq[0] - 1)
			case 'E': // Cursor Next Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.shiftCursorY(seq[0])
				v.setCursorX(0)
			case 'F': // Cursor Previous Line
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.setCursorX(0)
				v.shiftCursorY(-seq[0])
			case 'G': // Cursor Horizontal Absolute
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.setCursorX(seq[0] - 1)
			case 'H', 'f': // Cursor Position
				seq := parseSemicolonNumSeq(parameterCode, 1)
				if parameterCode == "" {
					v.setCursorPos(0, 0)
				} else {
					v.setCursorY(seq[0] - 1)
					if len(seq) > 1 {
						v.setCursorX(seq[1] - 1)
					}
				}
			case 'J': // Erase in Display
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from Cursor to end of screen
					for i := v.Cursor.X; i < len(v.screen[v.Cursor.Y]); i++ {
						v.screen[v.Cursor.Y][i].Rune = ' '
					}
					if v.Cursor.Y+1 < len(v.screen) {
						for j := v.Cursor.Y; j < len(v.screen); j++ {
							for i := 0; i < len(v.screen[j]); i++ {
								v.screen[j][i].Rune = ' '
							}
						}
					}
					v.RedrawWindow()
				case 1: // clear from Cursor to beginning of screen
					for j := 0; j < v.Cursor.Y; j++ {
						for i := 0; i < len(v.screen[j]); j++ {
							v.screen[j][i].Rune = ' '
						}
					}
					v.RedrawWindow()
				case 2: // clear entire screen (and move Cursor to top left?)
					for i := range v.screen {
						for j := range v.screen[i] {
							v.screen[i][j].Rune = ' '
						}
					}
					v.setCursorPos(0, 0)
					v.RedrawWindow()
				case 3: // clear entire screen and delete all lines saved in scrollback buffer
					v.scrollback = [][]render.Char{}
					for i := range v.screen {
						for j := range v.screen[i] {
							v.screen[i][j].Rune = ' '
						}
					}
					v.setCursorPos(0, 0)
					v.RedrawWindow()
				}
			case 'K': // Erase in Line
				seq := parseSemicolonNumSeq(parameterCode, 0)
				switch seq[0] {
				case 0: // clear from Cursor to end of line
					for i := v.Cursor.X; i < len(v.screen[v.Cursor.Y]); i++ { // FIXME: sometimes crashes
						v.screen[v.Cursor.Y][i].Rune = ' '
					}
				case 1: // clear from Cursor to beginning of line
					for i := 0; i < v.Cursor.X; i++ {
						v.screen[v.Cursor.Y][i].Rune = ' '
					}
				case 2: // clear entire line; Cursor position remains the same
					for i := 0; i < len(v.screen[v.Cursor.Y]); i++ {
						v.screen[v.Cursor.Y][i].Rune = ' '
					}
				}
				v.RedrawWindow()
			case 'r': // Set Scrolling Region
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.scrollingRegion.top = seq[0] - 1
				if len(seq) > 1 {
					v.scrollingRegion.bottom = seq[1] - 1
				} else {
					v.scrollingRegion.bottom = v.h + 1
				}
				v.setCursorPos(0, 0)
			case 'S': // Scroll Up; new lines added to bottom
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.scrollUp(seq[0])
			case 'T': // Scroll Down; new lines added to top
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.scrollDown(seq[0])
			case 'L': // Insert Lines
				seq := parseSemicolonNumSeq(parameterCode, 1)
				v.setCursorX(0)

				n := seq[0]
				newLines := make([][]render.Char, n)
				for i := range newLines {
					newLines[i] = make([]render.Char, v.w)
				}

				v.screen = append(append(
					newLines,
					v.screen[v.Cursor.Y:v.scrollingRegion.bottom-n+1]...),
					v.screen[v.scrollingRegion.bottom+1:]...)

				v.RedrawWindow()
			case 'm': // Select Graphic Rendition
				v.handleSDR(parameterCode)
			case 's': // Save Cursor Position
				v.storedCursorX = v.Cursor.X
				v.storedCursorY = v.Cursor.Y
			case 'u': // Restore Cursor Positon
				v.setCursorPos(v.storedCursorX, v.storedCursorY)
			default:
				v.renderer.Debug(string(next) + parameterCode)
				// v.debug("CSI Code: " + string(next) + " ; " + parameterCode)
			}
			return
		}
	}
}
