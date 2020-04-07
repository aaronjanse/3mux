package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aaronjanse/i3-tmux/render"
	"github.com/aaronjanse/i3-tmux/vterm"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	vterm *vterm.VTerm
	shell Shell

	searchMode bool
	searchText string

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

func (t *Pane) handleStdin(in string) {
	if t.searchMode {
		for _, c := range in {
			if c == 13 {
				continue // carriage return
			}
			if c == 8 || c == 127 { // backspace
				if len(t.searchText) > 0 {
					t.searchText = t.searchText[:len(t.searchText)-1]
				}
			} else if c == 10 {
				// TODO (newline)
			} else {
				t.searchText += string(c)
			}
		}
		t.displayStatusText(t.searchText)
	} else {
		t.shell.handleStdin(in)
	}
}

func (t *Pane) toggleSearch() {
	t.searchMode = !t.searchMode
	t.vterm.ChangePause <- t.searchMode

	if t.searchMode {
		// FIXME hacky way to wait for full control of screen section
		timer := time.NewTimer(time.Millisecond * 5)
		select {
		case <-timer.C:
			timer.Stop()
		}

		t.displayStatusText("Search...")
	} else {
		t.clearStatusText()
	}
}

func (t *Pane) displayStatusText(s string) {
	for i := 0; i < t.renderRect.w; i++ {
		r := ' '
		if i < len(s) {
			r = rune(s[i])
		}

		ch := render.PositionedChar{
			Rune: r,
			Cursor: render.Cursor{
				X: i,
				Y: t.renderRect.h - 1,
				Style: render.Style{
					Bg: render.Color{
						ColorMode: render.ColorBit3Bright,
						Code:      2,
					},
					Fg: render.Color{
						ColorMode: render.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		renderer.ForceHandleCh(ch)
	}
}

func (t *Pane) clearStatusText() {
	for i := 0; i < t.renderRect.w; i++ {
		ch := render.PositionedChar{
			Rune: ' ',
			Cursor: render.Cursor{
				X: i,
				Y: t.renderRect.h - 1,
				Style: render.Style{
					Bg: render.Color{
						ColorMode: render.ColorBit3Bright,
						Code:      2,
					},
					Fg: render.Color{
						ColorMode: render.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		renderer.ForceHandleCh(ch)
	}
}

func (t *Pane) kill() {
	t.vterm.Kill()
	t.shell.Kill()
}

func (t *Pane) setPause(pause bool) {
	t.vterm.ChangePause <- pause
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
