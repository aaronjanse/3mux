package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
	"github.com/kr/pty"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	id int

	selected bool

	renderRect Rect

	ptmx *os.File
	cmd  *exec.Cmd

	vterm *vterm.VTerm

	vtermOut <-chan vterm.Char
}

func (t *Pane) serialize() string {
	return fmt.Sprintf("Term")
}

func (t *Pane) setRenderRect(x, y, w, h int) {
	t.renderRect = Rect{x, y, w, h}

	t.softRefresh()

	t.vterm.Reshape(w, h)

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			err := pty.Setsize(t.ptmx, &pty.Winsize{
				Rows: uint16(h), Cols: uint16(w),
				X: 16 * uint16(w), Y: 16 * uint16(h),
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

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
