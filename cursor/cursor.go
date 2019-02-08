package cursor

// Cursor is the state of the terminal's drawing modes when printing a given character
type Cursor struct {
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
