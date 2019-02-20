package render

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

// Color stores sufficient data to reproduce an ANSI-encodable color
type Color struct {
	ColorMode
	Code int32
}

// ToANSI emits an ANSI SGR escape code for a given color
func (c Color) ToANSI(bg bool) string {
	var offset int32
	if bg {
		offset = 10
	} else {
		offset = 0
	}

	switch c.ColorMode {
	case ColorNone:
		return fmt.Sprintf("\033[%dm", 39+offset)
	case ColorBit3Normal:
		return fmt.Sprintf("\033[%dm", 30+offset+c.Code)
	case ColorBit3Bright:
		return fmt.Sprintf("\033[%dm", 90+offset+c.Code)
	case ColorBit8:
		return fmt.Sprintf("\033[%d;5;%dm", 38+offset, c.Code)
	case ColorBit24:
		return fmt.Sprintf(
			"\033[%d;2;%d;%d;%dm", 38+offset,
			(c.Code>>16)&0xff, (c.Code>>8)&0xff, c.Code&0xff,
		)
	default:
		panic(fmt.Sprintf("Unexpected ColorMode: %v", c.ColorMode))
	}
}
