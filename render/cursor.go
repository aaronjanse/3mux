package render

import (
	"fmt"

	"github.com/aaronjanse/3mux/ecma48"
)

// deltaMarkup returns markup to transform from one cursor to another
func deltaMarkup(fromCur, toCur ecma48.Cursor) string {
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

	differentBgCode := (to.Bg.ColorMode != ecma48.ColorNone) && to.Bg.Code != from.Bg.Code
	if to.Bg.ColorMode != from.Bg.ColorMode || differentBgCode {
		out += to.Bg.ToANSI(true)
	}

	differentFgCode := (to.Fg.ColorMode != ecma48.ColorNone) && to.Fg.Code != from.Fg.Code
	if to.Fg.ColorMode != from.Fg.ColorMode || differentFgCode {
		out += to.Fg.ToANSI(false)
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
