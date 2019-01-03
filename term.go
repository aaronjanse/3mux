package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
	"github.com/kr/pty"
)

func (t *Term) setRenderRect(x, y, w, h int) {
	t.renderRect = Rect{x, y, w, h}

	t.softRefresh()

	t.vterm.Reshape(w, h)
	t.vterm.RedrawWindow()

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
}

func (t *Term) softRefresh() {
	t.vterm.Selected = t.selected

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
		for i := 0; i < r.h+1; i++ {
			// fmt.Print(ansi.MoveTo(r.x-1, r.y+i) + borderCol + "│\033[0m")
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
		for i := 0; i < r.h+1; i++ {
			// fmt.Print(ansi.MoveTo(r.x+r.w, r.y+i) + borderCol + "│\033[0m")
			globalCharAggregate <- vterm.Char{
				Rune: '│',
				Cursor: cursor.Cursor{
					X: r.x + r.w + 1,
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
		// fmt.Print(ansi.MoveTo(r.x, r.y-1) + borderCol + strings.Repeat("─", r.w) + "\033[0m")
		for i := 0; i < r.w+1; i++ {
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
		// fmt.Print(ansi.MoveTo(r.x, r.y+r.h) + borderCol + strings.Repeat("─", r.w) + "\033[0m")
		for i := 0; i < r.w+1; i++ {
			globalCharAggregate <- vterm.Char{
				Rune: '─',
				Cursor: cursor.Cursor{
					X: r.x + i,
					Y: r.y + r.h + 1,
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
		// fmt.Print(ansi.MoveTo(r.x-1, r.y-1) + borderCol + "┌\033[0m")
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
		// fmt.Print(ansi.MoveTo(r.x+r.w, r.y-1) + borderCol + "┐\033[0m")
		globalCharAggregate <- vterm.Char{
			Rune: '┐',
			Cursor: cursor.Cursor{
				X: r.x + r.w + 1,
				Y: r.y - 1,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
	if bottomBorder && leftBorder {
		// fmt.Print(ansi.MoveTo(r.x-1, r.y+r.h) + borderCol + "└\033[0m")
		globalCharAggregate <- vterm.Char{
			Rune: '└',
			Cursor: cursor.Cursor{
				X: r.x - 1,
				Y: r.y + r.h + 1,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
	if bottomBorder && rightBorder {
		// fmt.Print(ansi.MoveTo(r.x+r.w, r.y+r.h) + borderCol + "┘\033[0m")
		globalCharAggregate <- vterm.Char{
			Rune: '┘',
			Cursor: cursor.Cursor{
				X: r.x + r.w + 1,
				Y: r.y + r.h + 1,
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      6,
				},
			},
		}
	}
}
