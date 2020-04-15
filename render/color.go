package render

import (
	"fmt"

	"github.com/aaronjanse/3mux/ecma48"
)

// ToANSI emits an ANSI SGR escape code for a given color
func ToANSI(c ecma48.Color, bg bool) string {
	var offset int32
	if bg {
		offset = 10
	} else {
		offset = 0
	}

	switch c.ColorMode {
	case ecma48.ColorNone:
		return fmt.Sprintf("\033[%dm", 39+offset)
	case ecma48.ColorBit3Normal:
		return fmt.Sprintf("\033[%dm", 30+offset+c.Code)
	case ecma48.ColorBit3Bright:
		return fmt.Sprintf("\033[%dm", 90+offset+c.Code)
	case ecma48.ColorBit8:
		return fmt.Sprintf("\033[%d;5;%dm", 38+offset, c.Code)
	case ecma48.ColorBit24:
		return fmt.Sprintf(
			"\033[%d;2;%d;%d;%dm", 38+offset,
			(c.Code>>16)&0xff, (c.Code>>8)&0xff, c.Code&0xff,
		)
	default:
		panic(fmt.Sprintf("Unexpected ColorMode: %v", c.ColorMode))
	}
}
