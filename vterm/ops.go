package vterm

import (
	"github.com/aaronjanse/i3-tmux/render"
)

func (v *VTerm) ScrollbackReset() {
	v.scrollbackPos = 0

	v.RedrawWindow()
}

// ScrollbackUp shifts the screen contents up, with scrollback
func (v *VTerm) ScrollbackUp() {
	if v.usingAltScreen {
		return
	}

	if v.scrollbackPos-5 >= 0 {
		v.scrollbackPos -= 5
	}

	v.RedrawWindow()
}

// ScrollbackDown shifts the screen contents down, with scrollback
func (v *VTerm) ScrollbackDown() {
	if v.usingAltScreen {
		return
	}

	if len(v.scrollback) == 0 {
		return
	}

	if v.scrollbackPos+5 < len(v.scrollback) {
		v.scrollbackPos += 5
		v.RedrawWindow()
	}
}

// RefreshCursor refreshes the ncurses cursor position
func (v *VTerm) RefreshCursor() {
	if v.isPaused {
		return
	}
	v.parentSetCursor(v.Cursor.X, v.Cursor.Y)
}

// scrollUp shifts screen contents up and adds blank lines to the bottom of the screen.
// Lines pushed out of view are put in the scrollback.
func (v *VTerm) scrollUp(n int) {
	if !v.usingAltScreen {
		rows := v.screen[v.scrollingRegion.top : v.scrollingRegion.top+n]
		v.scrollback = append(v.scrollback, rows...)
	}

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

	if !v.usingSlowRefresh {
		v.RedrawWindow()
	}
}

// scrollDown shifts the screen content down and adds blank lines to the top.
// It does neither modifies nor reads scrollback
func (v *VTerm) scrollDown(n int) {
	newLines := make([][]render.Char, n)
	for i := range newLines {
		newLines[i] = v.blankLine
	}

	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		newLines...),
		v.screen[v.scrollingRegion.top:v.scrollingRegion.bottom-n]...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	if !v.usingSlowRefresh {
		v.RedrawWindow()
	}
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

	positionedChar.Cursor.X += v.x
	positionedChar.Cursor.Y += v.y

	if v.Cursor.Y < len(v.screen) && v.Cursor.X < len(v.screen[v.Cursor.Y]) {
		v.screen[v.Cursor.Y][v.Cursor.X] = char
	}

	// TODO: print to the window based on scrolling position
	v.renderer.HandleCh(positionedChar)

	if v.Cursor.X < v.w {
		v.Cursor.X++
	}

	v.RefreshCursor()
}

// RedrawWindow redraws the screen into ncurses from scratch.
// This should be reserved for operations not yet formalized into a generic, efficient function.
func (v *VTerm) RedrawWindow() {
	if v.isPaused {
		return
	}
	if v.scrollbackPos < v.h {
		for y := 0; y < v.h-v.scrollbackPos; y++ {
			for x := 0; x < v.w; x++ {
				if y >= len(v.screen) || x >= len(v.screen[y]) {
					continue
				}

				ch := render.PositionedChar{
					Rune: v.screen[y][x].Rune,
					Cursor: render.Cursor{
						X: v.x + x, Y: v.y + y + v.scrollbackPos, Style: v.screen[y][x].Style,
					},
				}

				v.renderer.HandleCh(ch)
			}
		}
	}

	if !v.usingSlowRefresh {
		v.RefreshCursor()
	}

	if v.scrollbackPos > 0 {
		numLinesVisible := v.scrollbackPos
		if v.scrollbackPos > v.h {
			numLinesVisible = v.h
		}
		for y := 0; y < numLinesVisible; y++ {
			for x := 0; x < v.w; x++ {
				idx := len(v.scrollback) - v.scrollbackPos + y - 1

				if x < len(v.scrollback[idx]) {
					ch := render.PositionedChar{
						Rune: v.scrollback[idx][x].Rune,
						Cursor: render.Cursor{
							X: v.x + x, Y: v.y + y, Style: v.scrollback[idx][x].Style,
						},
					}
					v.renderer.HandleCh(ch)
				} else {
					ch := render.PositionedChar{
						Rune: ' ',
						Cursor: render.Cursor{
							X: v.x + x, Y: v.y + y, Style: render.Style{},
						},
					}
					v.renderer.HandleCh(ch)
				}
			}
		}
	}
}
