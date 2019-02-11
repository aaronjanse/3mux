package main

import (
	"fmt"
	"math/rand"

	"github.com/aaronduino/i3-tmux/render"
	"github.com/aaronduino/i3-tmux/vterm"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	vterm *vterm.VTerm
	shell Shell
}

func newTerm(selected bool) *Pane {
	stdout := make(chan rune, 32)
	shell := newShell(stdout)

	vtermOut := make(chan render.PositionedChar, 32)

	t := &Pane{
		id:       rand.Intn(10),
		selected: selected,

		shell: shell,
	}

	parentSetCursor := func(x, y int) {
		if t.selected {
			renderer.SetCursor(x+t.renderRect.x, y+t.renderRect.y)
		}
	}

	vt := vterm.NewVTerm(renderer, parentSetCursor, stdout, vtermOut)
	go vt.ProcessStream()

	t.vterm = vt

	go (func() {
		for {
			select {
			case char := <-vtermOut:
				if char.Cursor.X > t.renderRect.w-1 {
					continue
				}
				if char.Cursor.Y > t.renderRect.h-1 {
					continue
				}
				char.Cursor.X += t.renderRect.x
				char.Cursor.Y += t.renderRect.y
				renderer.RenderQueue <- char
			}
		}
	})()

	return t
}

func (t *Pane) kill() {
	t.vterm.Kill()
	t.shell.Kill()
}

func (t *Pane) serialize() string {
	return fmt.Sprintf("Term")
}

func (t *Pane) setRenderRect(x, y, w, h int) {
	// r := t.renderRect
	// if x == r.x && y == r.y && w == r.w && h == r.h {
	// 	return
	// }

	t.renderRect = Rect{x, y, w, h}

	t.vterm.Reshape(w, h)
	t.vterm.RedrawWindow()
	// renderer.Refresh()

	t.shell.resize(w, h)

	t.softRefresh()
}

func (t *Pane) softRefresh() {
	if t.selected {
		drawSelectionBorder(t.renderRect)
		// t.win.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	}
}
