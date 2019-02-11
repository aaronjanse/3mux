package render

import "fmt"

// Cursor is Style along with position
type Cursor struct {
	X, Y int
	Style
}

// Style is the state of the terminal's drawing modes when printing a given character
type Style struct {
	Bold, Faint, Italic, Underline, Conceal, CrossedOut bool

	// Fg is the foreground color
	Fg Color
	// Bg is the background color
	Bg Color
}

// Reset sets all rendering attributes of a cursor to their default values
func (c *Cursor) Reset() {
	c.Bold = false
	c.Faint = false
	c.Italic = false
	c.Underline = false
	c.Conceal = false
	c.CrossedOut = false

	c.Fg.ColorMode = ColorNone
	c.Bg.ColorMode = ColorNone
}

// deltaMarkup returns markup to transform from one cursor to another
func deltaMarkup(fromCur, toCur Cursor) string {
	out := ""

	xDiff := toCur.X - fromCur.X
	yDiff := toCur.Y - fromCur.Y

	if yDiff == 0 {
		if xDiff > 0 {
			out += fmt.Sprintf("\033[%dC", xDiff) // move forwards
		} else {
			out += fmt.Sprintf("\033[%dD", -xDiff) // move backwards
		}
	} else {
		out += fmt.Sprintf("\033[%d;%dH", toCur.Y, toCur.X)
	}

	out += fmt.Sprintf("\033[%d;%dH", toCur.Y+1, toCur.X+1)

	to := toCur.Style
	from := fromCur.Style

	if to.Bg.ColorMode != from.Bg.ColorMode || to.Bg.Code != from.Bg.Code {
		out += to.Bg.ToANSI(true)
	}

	if to.Fg.ColorMode != from.Fg.ColorMode || to.Fg.Code != from.Fg.Code {
		out += to.Fg.ToANSI(false)
	}

	/* removing effects */

	if from.Faint && !to.Faint {
		out += "\033[22m"
	}

	if from.Underline && !to.Underline {
		out += "\033[24m"
	}

	/* adding effects */

	if !from.Faint && to.Faint {
		out += "\033[2m"
	}

	if !from.Underline && to.Underline {
		out += "\033[4m"
	}

	return out
}
