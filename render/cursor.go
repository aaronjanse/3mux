package render

import (
	"fmt"

	"github.com/aaronjanse/3mux/ecma48"
)

// Cursor is Style along with position. Coordinates are relative to top left
type Cursor struct {
	X, Y int
	Style
}

// Style is the state of the terminal's drawing modes when printing a given character
type Style struct {
	Bold, Faint, Italic, Underline, Conceal, CrossedOut, Reverse bool

	Fg ecma48.Color // foreground color
	Bg ecma48.Color // background color
}

// Reset sets all rendering attributes of a cursor to their default values
func (s *Style) Reset() {
	s.Bold = false
	s.Faint = false
	s.Italic = false
	s.Underline = false
	s.Conceal = false
	s.CrossedOut = false
	s.Reverse = false

	s.Fg.ColorMode = ecma48.ColorNone
	s.Bg.ColorMode = ecma48.ColorNone
}

// deltaMarkup returns markup to transform from one cursor to another
func deltaMarkup(fromCur, toCur Cursor) string {
	out := ""

	/* update position */

	xDiff := toCur.X - fromCur.X
	yDiff := toCur.Y - fromCur.Y

	if xDiff == 0 && yDiff == 1 {
		out += "\033[1B"
	} else if xDiff != 0 || yDiff != 0 {
		out += fmt.Sprintf("\033[%d;%dH", toCur.Y+1, toCur.X+1)
	}

	/* update colors */

	to := toCur.Style
	from := fromCur.Style

	if to.Bg.ColorMode != from.Bg.ColorMode || to.Bg.Code != from.Bg.Code {
		out += ToANSI(to.Bg, true)
	}

	if to.Fg.ColorMode != from.Fg.ColorMode || to.Fg.Code != from.Fg.Code {
		out += ToANSI(to.Fg, false)
	}

	/* remove effects */

	if from.Bold && !to.Bold {
		out += "\033[22m"
	}

	if from.Faint && !to.Faint {
		out += "\033[22m"
	}

	if from.Underline && !to.Underline {
		out += "\033[24m"
	}

	if from.Reverse && !to.Reverse {
		out += "\033[27m"
	}

	/* add effects */

	if !from.Bold && to.Bold {
		out += "\033[1m"
	}

	if !from.Faint && to.Faint {
		out += "\033[2m"
	}

	if !from.Underline && to.Underline {
		out += "\033[4m"
	}

	if !from.Reverse && to.Reverse {
		out += "\033[7m"
	}

	return out
}
