package cursor

import "fmt"

// ColorMode is the type of color associated with a cursor
type ColorMode int

const (
	// ColorNone is the default unset color state
	ColorNone ColorMode = iota
	// ColorBit3Normal is for the 8 default non-bright colors
	ColorBit3Normal
	// ColorBit3Bright is for the 8 default bright colors
	ColorBit3Bright
	// ColorBit8 is specified at https://en.wikipedia.org/w/index.php?title=ANSI_escape_code&oldid=873901864#8-bit
	ColorBit8
	// ColorBit24 is specified at https://en.wikipedia.org/w/index.php?title=ANSI_escape_code&oldid=873901864#24-bit
	ColorBit24
)

// Cursor is the state of the terminal's drawing modes when printing a given character
type Cursor struct {
	Bold, Faint, Italic, Underline, Conceal, CrossedOut bool

	ColorMode
	Color int

	X, Y int
}

func (c Cursor) toMarkup() {

}

// DeltaMarkup returns markup to transform from one cursor to another
func DeltaMarkup(from, to Cursor) string {
	out := ""

	xDiff := to.X - from.X
	yDiff := to.Y - from.Y

	if yDiff == 0 {
		if xDiff > 0 {
			out += fmt.Sprintf("\033[%dC", xDiff) // move forwards
		} else {
			out += fmt.Sprintf("\033[%dD", -xDiff) // move backwards
		}
	} else {
		out += fmt.Sprintf("\033[%d;%dH", to.Y, to.X)
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
}
