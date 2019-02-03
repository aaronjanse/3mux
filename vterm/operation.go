package vterm

import (
	"fmt"
	"strings"

	"github.com/aaronduino/i3-tmux/cursor"
)

/*
The idea behind "operations" is that sometimes a virtual terminal needs low-level control of the host terminal.

For example:
- scrolling can be optimized via ephemeral scrolling regions
- images
- bracketed paste
- other input mode stuff
*/

// Operation is an action that can be encoded into ansi given the proper parameters from the window manager
type Operation interface {
	Serialize(x, y, w, h int, c cursor.Cursor) string
}

// ScrollDown moves the text of a pane up, adding new lines to the bottom
type ScrollDown struct{ numLines int }

// Serialize turns the ScrollDown operation into ansi codes
func (oper ScrollDown) Serialize(x, y, w, h int, c cursor.Cursor) string {
	out := fmt.Sprintf("\033[%v;%vr", y, y+h) // set top/bottom margins
	out += "\033[?69h"                        // enable left/right margins
	out += fmt.Sprintf("\033[%v;%vs", x, x+w) // set left/right margins

	out += fmt.Sprintf("\033[%v;%vH", y+h, x+w-1) // move into position to scroll
	out += strings.Repeat("\n", oper.numLines)    // print newlines to scroll

	out += "\033[s"    // reset left/right margins
	out += "\033[?69l" // disable left/right margins
	out += "\033[r"    // reset top/bottom margins
	out += fmt.Sprintf("\033[%v;%vH", c.Y, c.X)

	return out
}
