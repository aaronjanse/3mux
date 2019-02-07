package main

import (
	"fmt"
	"math/rand"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	shell Shell

	vterm *vterm.VTerm

	vtermOut <-chan vterm.Char
}

func newTerm(selected bool) *Pane {
	stdout := make(chan rune, 32)
	shell := newShell(stdout)

	vtermOut := make(chan vterm.Char, 32)

	vt := vterm.NewVTerm(stdout, vtermOut)
	go vt.ProcessStream()

	t := &Pane{
		id:       rand.Intn(10),
		selected: selected,

		shell:    shell,
		vterm:    vt,
		vtermOut: vtermOut,
	}

	go (func() {
		for {
			char := <-vtermOut
			if char.Cursor.X > t.renderRect.w-1 {
				continue
			}
			if char.Cursor.Y > t.renderRect.h-1 {
				continue
			}
			char.Cursor.X += t.renderRect.x
			char.Cursor.Y += t.renderRect.y
			globalCharAggregate <- char
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
	t.renderRect = Rect{x, y, w, h}

	t.softRefresh()

	t.vterm.Reshape(w, h)
	t.shell.resize(w, h)

	t.vterm.RedrawWindow()
}

func (t *Pane) softRefresh() {
	if t.selected {
		drawSelectionBorder(t.renderRect)
	}
}

func drawSelectionBorder(r Rect) {
	leftBorder := r.x > 0
	rightBorder := r.x+r.w+1 < termW
	topBorder := r.y > 0
	bottomBorder := r.y+r.h+1 < termH

	// draw lines
	if leftBorder {
		for i := 0; i <= r.h; i++ {
			globalCharAggregate <- vterm.Char{
				Rune: '│',
				Cursor: cursor.Cursor{
					X: r.x - 1,
					Y: r.y + i,
					Fg: cursor.Color{
						ColorMode: cursor.ColorBit3Normal,
						Code:      6,
					},
				},
			}
		}
	}
	if rightBorder {
		for i := 0; i <= r.h; i++ {
			globalCharAggregate <- vterm.Char{
				Rune: '│',
				Cursor: cursor.Cursor{
					X: r.x + r.w,
					Y: r.y + i,
					Fg: cursor.Color{
						ColorMode: cursor.ColorBit3Normal,
						Code:      6,
					},
				},
			}
		}
	}
	if topBorder {
		for i := 0; i <= r.w; i++ {
			globalCharAggregate <- vterm.Char{
				Rune: '─',
				Cursor: cursor.Cursor{
					X: r.x + i,
					Y: r.y - 1,
					Fg: cursor.Color{
						ColorMode: cursor.ColorBit3Normal,
						Code:      6,
					},
				},
			}
		}
	}
	if bottomBorder {
		for i := 0; i <= r.w; i++ {
			globalCharAggregate <- vterm.Char{
				Rune: '─',
				Cursor: cursor.Cursor{
					X: r.x + i,
					Y: r.y + r.h,
					Fg: cursor.Color{
						ColorMode: cursor.ColorBit3Normal,
						Code:      6,
					},
				},
			}
		}
	}

	// draw corners
	if topBorder && leftBorder {
		globalCharAggregate <- vterm.Char{
			Rune: '┌',
			Cursor: cursor.Cursor{
				X: r.x - 1,
				Y: r.y - 1,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
	if topBorder && rightBorder {
		globalCharAggregate <- vterm.Char{
			Rune: '┐',
			Cursor: cursor.Cursor{
				X: r.x + r.w,
				Y: r.y - 1,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
	if bottomBorder && leftBorder {
		globalCharAggregate <- vterm.Char{
			Rune: '└',
			Cursor: cursor.Cursor{
				X: r.x - 1,
				Y: r.y + r.h,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
	if bottomBorder && rightBorder {
		globalCharAggregate <- vterm.Char{
			Rune: '┘',
			Cursor: cursor.Cursor{
				X: r.x + r.w,
				Y: r.y + r.h,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
}
