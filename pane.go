package main

import (
	"fmt"
	"math/rand"

	"github.com/aaronjanse/i3-tmux/vterm"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	vterm *vterm.VTerm
	shell Shell

	Dead bool
}

func newTerm(selected bool) *Pane {
	stdout := make(chan rune, 3200000)

	t := &Pane{
		id:       rand.Intn(10),
		selected: selected,

		shell: newShell(stdout),
	}

	go func() {
		t.shell.cmd.Wait()
		t.Dead = true
		removeTheDead([]int{root.selectionIdx})

		if len(root.workspaces[root.selectionIdx].contents.elements) == 0 {
			shutdownNow()
		} else {
			// deselect the old Term
			newTerm := getSelection().getContainer().(*Pane)
			newTerm.selected = true
			newTerm.softRefresh()
			newTerm.vterm.RefreshCursor()

			root.simplify()
			root.refreshRenderRect()
		}
	}()

	parentSetCursor := func(x, y int) {
		if t.selected {
			renderer.SetCursor(x+t.renderRect.x, y+t.renderRect.y)
		}
	}

	vt := vterm.NewVTerm(&t.shell.byteCounter, renderer, parentSetCursor, stdout)
	go vt.ProcessStream()

	t.vterm = vt

	return t
}

func (t *Pane) kill() {
	t.vterm.Kill()
	t.shell.Kill()
}

func (t *Pane) setPause(pause bool) {
	t.vterm.Paused = pause
}

func (t *Pane) serialize() string {
	return fmt.Sprintf("Term")
}

func (t *Pane) simplify() {}

func (t *Pane) setRenderRect(x, y, w, h int) {
	t.renderRect = Rect{x, y, w, h}

	t.vterm.Reshape(x, y, w, h)
	t.vterm.RedrawWindow()

	t.shell.resize(w, h)

	t.softRefresh()
}

func (t *Pane) getRenderRect() Rect {
	return t.renderRect
}

func (t *Pane) softRefresh() {
	// only selected Panes get the special highlight color
	if t.selected {
		drawSelectionBorder(t.renderRect)
	}
}
