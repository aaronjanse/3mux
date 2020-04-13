package vterm

// Renderer is how vterm displays its output
type Renderer interface {
	SetChar(ch Char, x, y int)
	Refresh()
}

type Char struct {
	Rune rune
	Style
}

type Cursor struct {
	X, Y int
	Style
}

// Style is the state of the terminal's drawing modes when printing a given character
type Style struct {
	Bold, Faint, Italic, Underline, Conceal, CrossedOut bool

	Fg Color // foreground color
	Bg Color // background color
}

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

// Color stores sufficient data to reproduce an ANSI-encodable color
type Color struct {
	ColorMode
	Code int32
}
