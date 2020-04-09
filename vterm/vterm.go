/*
Package vterm provides a layer of abstraction between a channel of incoming text (possibly containing ANSI escape codes, et al) and a channel of outbound Char's.

A Char is a character printed using a given cursor (which is stored alongside the Char).
*/
package vterm

import (
	"github.com/aaronjanse/i3-tmux/render"
)

// ScrollingRegion holds the state for an ANSI scrolling region
type ScrollingRegion struct {
	top    int
	bottom int
}

/*
VTerm acts as a virtual terminal emulator between a shell and the host terminal emulator

It both transforms an inbound stream of bytes into Char's and provides the option of dumping all the Char's that need to be rendered to display the currently visible terminal window from scratch.
*/
type VTerm struct {
	x, y, w, h int

	// visible screen; char cursor coords are ignored
	Screen [][]render.Char

	// Scrollback[0] is the line farthest from the screen
	Scrollback    [][]render.Char // disabled when using alt screen; char cursor coords are ignored
	ScrollbackPos int             // ScrollbackPos is the number of lines of scrollback visible

	usingAltScreen bool
	screenBackup   [][]render.Char

	NeedsRedraw bool

	startTime           int64
	shellByteCounter    *uint64
	internalByteCounter uint64
	usingSlowRefresh    bool

	Cursor render.Cursor

	renderer *render.Renderer

	// TODO: delete `blankLine`
	blankLine []render.Char

	// parentSetCursor sets physical host's cursor taking the pane location into account
	parentSetCursor func(x, y int)

	in  <-chan rune
	out chan<- rune

	storedCursorX, storedCursorY int

	scrollingRegion ScrollingRegion

	ChangePause   chan bool
	IsPaused      bool
	DebugSlowMode bool
}

// NewVTerm returns a VTerm ready to be used by its exported methods
func NewVTerm(shellByteCounter *uint64, renderer *render.Renderer, parentSetCursor func(x, y int), in <-chan rune, out chan<- rune) *VTerm {
	w := 10
	h := 10

	screen := [][]render.Char{}
	for j := 0; j < h; j++ {
		row := []render.Char{}
		for i := 0; i < w; i++ {
			row = append(row, render.Char{
				Rune:  ' ',
				Style: render.Style{},
			})
		}
		screen = append(screen, row)
	}

	v := &VTerm{
		x: 0, y: 0,
		w:                w,
		h:                h,
		blankLine:        []render.Char{},
		Screen:           screen,
		Scrollback:       [][]render.Char{},
		usingAltScreen:   false,
		Cursor:           render.Cursor{},
		in:               in,
		out:              out,
		shellByteCounter: shellByteCounter,
		usingSlowRefresh: false,
		renderer:         renderer,
		parentSetCursor:  parentSetCursor,
		scrollingRegion:  ScrollingRegion{top: 0, bottom: h - 1},
		NeedsRedraw:      false,
		ChangePause:      make(chan bool, 1),
		IsPaused:         false,
		DebugSlowMode:    false,
	}

	return v
}

// Kill safely shuts down all vterm processes for the instance
func (v *VTerm) Kill() {
	v.usingSlowRefresh = false
	v.ChangePause <- true
}

// Reshape safely updates a VTerm's width & height
func (v *VTerm) Reshape(x, y, w, h int) {

	v.x = x
	v.y = y

	for y := 0; y <= h; y++ {
		if y >= len(v.Screen) {
			v.Screen = append(v.Screen, []render.Char{})
		}

		for x := 0; x <= w; x++ {
			if x >= len(v.Screen[y]) {
				v.Screen[y] = append(v.Screen[y], render.Char{Rune: ' ', Style: render.Style{}})
			}
		}
	}

	if len(v.Screen)-1 > h {
		diff := len(v.Screen) - h - 1
		v.Scrollback = append(v.Scrollback, v.Screen[:diff]...)
		v.Screen = v.Screen[diff:]
	}

	if v.scrollingRegion.top == 0 && v.scrollingRegion.bottom == v.h-1 {
		v.scrollingRegion.bottom = h - 1
	}

	v.w = w
	v.h = h

	if v.Cursor.Y >= h {
		v.setCursorY(h - 1)
	}

	if v.Cursor.X >= w {
		v.setCursorX(w - 1)
	}

	v.RedrawWindow()
}
