package vterm

import (
	"github.com/aaronduino/i3-tmux/cursor"
	gc "github.com/rthornton128/goncurses"
)

// RefreshCursor refreshes the ncurses cursor position
func (v *VTerm) RefreshCursor() {
	v.win.Move(v.cursorY, v.cursorX)
	v.win.Refresh()
}

// scrollUp shifts screen contents up and adds blank lines to the bottom of the screen.
// Lines pushed out of view are put in the scrollback.
func (v *VTerm) scrollUp(n int) {
	// if !v.usingAltScreen {
	// 	v.scrollback = append(v.scrollback, v.screen[v.scrollingRegion.top:v.scrollingRegion.top+n]...)
	// }

	blankLine := []Char{}
	for i := 0; i < v.w; i++ {
		blankLine = append(blankLine, Char{Rune: ' '})
	}

	newLines := make([][]Char, n)
	for i := range newLines {
		newLines[i] = blankLine
	}

	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		v.screen[v.scrollingRegion.top+n:v.scrollingRegion.bottom+1]...),
		newLines...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	v.win.Scroll(n)
	// v.win.Refresh()
	// v.redrawWindow()
}

// scrollDown shifts the screen content down and adds blank lines to the top.
// It does neither modifies nor reads scrollback
func (v *VTerm) scrollDown(n int) {
	blankLine := []Char{}
	for i := 0; i < v.w; i++ {
		blankLine = append(blankLine, Char{Rune: ' '})
	}

	newLines := make([][]Char, n)
	for i := range newLines {
		newLines[i] = blankLine
	}

	// v.screen = append(v.screen[:len(v.screen)-1], []Char{})
	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		newLines...),
		v.screen[v.scrollingRegion.top:v.scrollingRegion.bottom-n]...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	// v.redrawWindow()
	v.win.Scroll(n)
}

func (v *VTerm) setCursorPos(x, y int) {
	// TODO: account for scrolling positon

	v.win.Move(y, x)

	v.cursorX = x
	v.cursorY = y
}

func (v *VTerm) setCursorX(x int) {
	v.setCursorPos(x, v.cursorY)
}

func (v *VTerm) setCursorY(y int) {
	v.setCursorPos(v.cursorX, y)
}

func (v *VTerm) shiftCursorX(diff int) {
	v.setCursorPos(v.cursorX+diff, v.cursorY)
}

func (v *VTerm) shiftCursorY(diff int) {
	v.setCursorPos(v.cursorX, v.cursorY+diff)
}

// putChar renders as given character using the cursor stored in vterm
func (v *VTerm) putChar(ch rune) {
	if v.cursorY >= v.h || v.cursorY < 0 || v.cursorX > v.w || v.cursorX < 0 {
		return
	}

	char := Char{
		Rune:   ch,
		Cursor: v.Cursor,
	}

	// fmt.Print(len(v.screen), v.cursorY, "")
	// fmt.Print(len(v.screen[v.cursorY]), v.cursorX, ",  ")

	if v.cursorY < len(v.screen) && v.cursorX < len(v.screen[v.cursorY]) {
		v.screen[v.cursorY][v.cursorX] = char
	}

	// TODO: set ncurses style attributes to match those of the cursor

	// TODO: print to the window based on scrolling position
	v.updateCursesStyle(v.Cursor)
	v.win.Print(string(ch))
	v.win.Refresh()

	if v.cursorX < v.w {
		v.cursorX++
	}
}

func (v *VTerm) updateCursesStyle(c cursor.Cursor) {
	cursorAttrs := []bool{
		v.Cursor.Bold,
		v.Cursor.Faint,
		v.Cursor.Underline,
	}
	cursesAttrs := []gc.Char{
		gc.A_BOLD,
		gc.A_DIM,
		gc.A_UNDERLINE,
	}
	for i := 0; i < len(cursorAttrs); i++ {
		if cursorAttrs[i] {
			v.win.AttrOn(cursesAttrs[i])
		} else {
			v.win.AttrOff(cursesAttrs[i])
		}
	}

	// gc.InitPair(1, gc.C_MAGENTA, -1)
	// gc.InitPair(2, gc.C_GREEN, -1)
	// gc.InitPair(3, gc.C_BLUE, -1)

	switch v.Cursor.Fg.ColorMode {
	case cursor.ColorNone:
		v.win.ColorOn(1)
	case cursor.ColorBit3Normal:
		v.win.ColorOn(int16(v.Cursor.Fg.Code) + 2)
	case cursor.ColorBit3Bright:
		v.win.ColorOn(int16(v.Cursor.Fg.Code) + 8 + 2)
	case cursor.ColorBit8:
		v.win.ColorOn(int16(v.Cursor.Fg.Code) + 2)
	case cursor.ColorBit24:
		code := v.Cursor.Fg.Code
		// r := int16((code >> 16) & 0xff)
		// g := int16((code >> 8) & 0xff)
		// b := int16(code & 0xff)
		// id := ((r + g + b) % 600) + 260
		// gc.InitColor(id, r, g, b)
		// gc.InitPair(id, id, -1)
		v.win.ColorOn(int16(code))
	}

	// v.win.Color(2)

}

// RedrawWindow redraws the screen into ncurses from scratch.
// This should be reserved for operations not yet formalized into a generic, efficient function.
func (v *VTerm) RedrawWindow() {
	for y := 0; y < v.h; y++ {
		for x := 0; x < v.w; x++ {
			v.updateCursesStyle(v.screen[y][x].Cursor)
			v.win.MovePrint(y, x, string(v.screen[y][x].Rune))
		}
	}
	v.win.Refresh()
}
