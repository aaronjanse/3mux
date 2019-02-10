package main

import (
	"fmt"
	"math/rand"

	"github.com/aaronduino/i3-tmux/vterm"
	gc "github.com/rthornton128/goncurses"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	vterm *vterm.VTerm
	shell Shell
	win   *gc.Window
}

func newTerm(selected bool) *Pane {
	stdout := make(chan rune, 32)
	shell := newShell(stdout)

	vtermOut := make(chan vterm.Char, 32)

	var win *gc.Window
	win, err := gc.NewWindow(10, 10, 0, 0)
	if err != nil {
		panic(err)
	}
	win.ScrollOk(true)

	vt := vterm.NewVTerm(win, stdout, vtermOut)
	go vt.ProcessStream()

	t := &Pane{
		id:       rand.Intn(10),
		selected: selected,

		shell: shell,
		vterm: vt,
		win:   win,
	}

	// go (func() {
	// 	for {
	// 		char := <-vtermOut
	// 		if char.Cursor.X > t.renderRect.w-1 {
	// 			continue
	// 		}
	// 		if char.Cursor.Y > t.renderRect.h-1 {
	// 			continue
	// 		}
	// 		char.Cursor.X += t.renderRect.x
	// 		char.Cursor.Y += t.renderRect.y
	// 		globalCharAggregate <- char
	// 	}
	// })()

	return t
}

func (t *Pane) kill() {
	t.vterm.Kill()
	t.shell.Kill()
	t.win.Erase()
	t.win.Refresh()
	t.win.Delete()
}

func (t *Pane) serialize() string {
	return fmt.Sprintf("Term")
}

func (t *Pane) setRenderRect(x, y, w, h int) {
	r := t.renderRect
	if x == r.x && y == r.y && w == r.w && h == r.h {
		return
	}

	t.renderRect = Rect{x, y, w, h}

	t.win.Resize(h, w)
	t.win.MoveWindow(y, x)

	t.vterm.Reshape(w, h)

	t.shell.resize(w, h)

	t.softRefresh()
}

func (t *Pane) softRefresh() {
	if t.selected {
		drawSelectionBorder(t.renderRect)
		// t.win.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	}
}
