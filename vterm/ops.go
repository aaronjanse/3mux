package vterm

import (
	"github.com/aaronduino/i3-tmux/render"
)

// RefreshCursor refreshes the ncurses cursor position
func (v *VTerm) RefreshCursor() {
	v.parentSetCursor(v.Cursor.X, v.Cursor.Y)
}

// scrollUp shifts screen contents up and adds blank lines to the bottom of the screen.
// Lines pushed out of view are put in the scrollback.
func (v *VTerm) scrollUp(n int) {
	// if !v.usingAltScreen {
	// 	v.scrollback = append(v.scrollback, v.screen[v.scrollingRegion.top:v.scrollingRegion.top+n]...)
	// }

	blankLine := []render.Char{}
	for i := 0; i < v.w; i++ {
		blankLine = append(blankLine, render.Char{Rune: ' ', Style: render.Style{}})
	}

	newLines := make([][]render.Char, n)
	for i := range newLines {
		newLines[i] = blankLine
	}

	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		v.screen[v.scrollingRegion.top+n:v.scrollingRegion.bottom+1]...),
		newLines...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	v.RedrawWindow() // FIXME
}

// scrollDown shifts the screen content down and adds blank lines to the top.
// It does neither modifies nor reads scrollback
func (v *VTerm) scrollDown(n int) {
	blankLine := []render.Char{}
	for i := 0; i < v.w; i++ {
		blankLine = append(blankLine, render.Char{Rune: ' ', Style: render.Style{}})
	}

	newLines := make([][]render.Char, n)
	for i := range newLines {
		newLines[i] = blankLine
	}

	// v.screen = append(v.screen[:len(v.screen)-1], []Char{})
	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		newLines...),
		v.screen[v.scrollingRegion.top:v.scrollingRegion.bottom-n]...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	v.RedrawWindow()
	// v.win.Scroll(n)
}

func (v *VTerm) setCursorPos(x, y int) {
	// TODO: account for scrolling positon

	v.Cursor.X = x
	v.Cursor.Y = y

	v.RefreshCursor()
}

func (v *VTerm) setCursorX(x int) {
	v.setCursorPos(x, v.Cursor.Y)
}

func (v *VTerm) setCursorY(y int) {
	v.setCursorPos(v.Cursor.X, y)
}

func (v *VTerm) shiftCursorX(diff int) {
	v.setCursorPos(v.Cursor.X+diff, v.Cursor.Y)
}

func (v *VTerm) shiftCursorY(diff int) {
	v.setCursorPos(v.Cursor.X, v.Cursor.Y+diff)
}

// putChar renders as given character using the cursor stored in vterm
func (v *VTerm) putChar(ch rune) {
	if v.Cursor.Y >= v.h || v.Cursor.Y < 0 || v.Cursor.X > v.w || v.Cursor.X < 0 {
		return
	}

	char := render.Char{
		Rune:  ch,
		Style: v.Cursor.Style,
	}

	positionedChar := render.PositionedChar{
		Rune:   ch,
		Cursor: v.Cursor,
	}

	// fmt.Print(len(v.screen), v.cursorY, "")
	// fmt.Print(len(v.screen[v.cursorY]), v.cursorX, ",  ")

	if v.Cursor.Y < len(v.screen) && v.Cursor.X < len(v.screen[v.Cursor.Y]) {
		v.screen[v.Cursor.Y][v.Cursor.X] = char
	}

	v.out <- positionedChar

	// // TODO: set ncurses style attributes to match those of the cursor

	// // TODO: print to the window based on scrolling position
	// v.updateCursesStyle(v.Cursor)
	// v.win.Print(string(ch))
	// v.win.Refresh()

	if v.Cursor.X < v.w {
		v.Cursor.X++
	}

	v.RefreshCursor()
}

// RedrawWindow redraws the screen into ncurses from scratch.
// This should be reserved for operations not yet formalized into a generic, efficient function.
func (v *VTerm) RedrawWindow() {
	for y := 0; y < v.h; y++ {
		for x := 0; x < v.w; x++ {
			v.out <- render.PositionedChar{
				Rune: v.screen[y][x].Rune,
				Cursor: render.Cursor{
					X: x, Y: y, Style: v.screen[y][x].Style,
				},
			}
		}
	}
	// v.renderer.Refresh()
}
