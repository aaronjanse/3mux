package cursor

import (
	"fmt"
)

// Cursor is the state of the terminal's drawing modes when printing a given character
type Cursor struct {
	Bold, Faint, Italic, Underline, Conceal, CrossedOut bool

	// Fg is the foreground color
	Fg Color
	// Bg is the background color
	Bg Color

	X, Y int
}

func (c Cursor) toMarkup() {

}

// DeltaMarkup returns markup to transform from one cursor to another
func DeltaMarkup(from, to Cursor) string {
	out := ""

	// xDiff := to.X - from.X
	// yDiff := to.Y - from.Y

	// if yDiff == 0 {
	// 	if xDiff > 0 {
	// 		out += fmt.Sprintf("\033[%dC", xDiff) // move forwards
	// 	} else {
	// 		out += fmt.Sprintf("\033[%dD", -xDiff) // move backwards
	// 	}
	// } else {
	// 	out += fmt.Sprintf("\033[%d;%dH", to.Y, to.X)
	// }

	out += fmt.Sprintf("\033[%d;%dH", to.Y+1, to.X+1)

	if to.Fg.ColorMode != from.Fg.ColorMode || to.Fg.Code != from.Fg.Code {
		out += to.Fg.ToANSI(false)
	}

	if to.Bg.ColorMode != from.Bg.ColorMode || to.Bg.Code != from.Bg.Code {
		out += to.Fg.ToANSI(true)
	}

	/* removing effects */

	if from.Faint && !to.Faint {
		out += "\033[22m"
	}

	/* adding effects */

	if !from.Faint && to.Faint {
		out += "\033[2m"
	}

	return out
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
