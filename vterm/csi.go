package vterm

import (
	"log"
	"unicode"

	"github.com/aaronjanse/i3-tmux/render"
)

func (v *VTerm) handleCSISequence() {
	var next rune
	parameterCode := ""
	privateSequence := false
	for {
		var ok bool
		next, ok = v.pullRune()
		if !ok {
			return
		}

		if unicode.IsDigit(next) || next == ';' || next == ' ' {
			parameterCode += string(next)
		} else if next == '?' {
			privateSequence = true
		} else {
			break
		}
	}

	if privateSequence {
		v.handlePrivateSequence(next, parameterCode)
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
			v.handleEraseInDisplay(parameterCode)
		case 'K': // Erase in Line
			v.handleEraseInLine(parameterCode)
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
		case 'L': // Insert Lines; https://vt100.net/docs/vt510-rm/IL.html
			seq := parseSemicolonNumSeq(parameterCode, 1)
			v.setCursorX(0)

			n := seq[0]
			newLines := make([][]render.Char, n)
			for i := range newLines {
				newLines[i] = make([]render.Char, v.w)
			}

			v.Screen = append(append(
				newLines,
				v.Screen[v.Cursor.Y:v.scrollingRegion.bottom-n+1]...),
				v.Screen[v.scrollingRegion.bottom+1:]...)

			v.RedrawWindow()
		case 'm': // Select Graphic Rendition
			v.handleSGR(parameterCode)
		case 's': // Save Cursor Position
			v.storedCursorX = v.Cursor.X
			v.storedCursorY = v.Cursor.Y
		case 'u': // Restore Cursor Positon
			v.setCursorPos(v.storedCursorX, v.storedCursorY)
		default:
			log.Printf("Unrecognized CSI Code: %v", parameterCode+string(next))
		}
	}
}

func (v *VTerm) handleEraseInDisplay(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)
	switch seq[0] {
	case 0: // clear from Cursor to end of screen
		for i := v.Cursor.X; i < len(v.Screen[v.Cursor.Y]); i++ {
			v.Screen[v.Cursor.Y][i].Rune = ' '
		}
		if v.Cursor.Y+1 < len(v.Screen) {
			for j := v.Cursor.Y; j < len(v.Screen); j++ {
				for i := 0; i < len(v.Screen[j]); i++ {
					v.Screen[j][i].Rune = ' '
				}
			}
		}
		v.RedrawWindow()
	case 1: // clear from Cursor to beginning of screen
		for j := 0; j < v.Cursor.Y; j++ {
			for i := 0; i < len(v.Screen[j]); j++ {
				v.Screen[j][i].Rune = ' '
			}
		}
		v.RedrawWindow()
	case 2: // clear entire screen (and move Cursor to top left?)
		for i := range v.Screen {
			for j := range v.Screen[i] {
				v.Screen[i][j].Rune = ' '
			}
		}
		v.setCursorPos(0, 0)
		v.RedrawWindow()
	case 3: // clear entire screen and delete all lines saved in scrollback buffer
		v.Scrollback = [][]render.Char{}
		for i := range v.Screen {
			for j := range v.Screen[i] {
				v.Screen[i][j].Rune = ' '
			}
		}
		v.setCursorPos(0, 0)
		v.RedrawWindow()
	default:
		log.Printf("Unrecognized erase in display directive: %v", seq[0])
	}
}

func (v *VTerm) handleEraseInLine(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)
	switch seq[0] {
	case 0: // clear from Cursor to end of line
		for i := v.Cursor.X; i < len(v.Screen[v.Cursor.Y]); i++ {
			v.Screen[v.Cursor.Y][i].Rune = ' '
		}
	case 1: // clear from Cursor to beginning of line
		for i := 0; i < v.Cursor.X; i++ {
			v.Screen[v.Cursor.Y][i].Rune = ' '
		}
	case 2: // clear entire line; Cursor position remains the same
		for i := 0; i < len(v.Screen[v.Cursor.Y]); i++ {
			v.Screen[v.Cursor.Y][i].Rune = ' '
		}
	default:
		log.Printf("Unrecognized erase in line directive: %v", seq[0])
	}
	v.RedrawWindow()
}
