package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aaronduino/i3-tmux/render"
)

func drawSelectionBorder(r Rect) {
	leftBorder := r.x > 0
	rightBorder := r.x+r.w+1 < termW
	topBorder := r.y > 0
	bottomBorder := r.y+r.h+1 < termH

	style := render.Style{
		Fg: render.Color{
			ColorMode: render.ColorBit3Normal,
			Code:      6,
		},
	}

	// draw lines
	if leftBorder {
		for i := 0; i <= r.h; i++ {
			// stdscr.MoveAddChar(r.y+i, r.x-1, gc.ACS_VLINE)
			renderer.RenderQueue <- render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.x - 1,
					Y:     r.y + i,
					Style: style,
				},
			}
		}
	}
	if rightBorder {
		for i := 0; i <= r.h; i++ {
			// stdscr.MoveAddChar(r.y+i, r.x+r.w, gc.ACS_VLINE)
			renderer.RenderQueue <- render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.x + r.w,
					Y:     r.y + i,
					Style: style,
				},
			}
		}
	}
	if topBorder {
		for i := 0; i <= r.w; i++ {
			// stdscr.MoveAddChar(r.y-1, r.x+i, gc.ACS_HLINE)
			renderer.RenderQueue <- render.PositionedChar{
				Rune: '─',
				Cursor: render.Cursor{
					X:     r.x + i,
					Y:     r.y - 1,
					Style: style,
				},
			}
		}
	}
	if bottomBorder {
		for i := 0; i <= r.w; i++ {
			// stdscr.MoveAddChar(r.y+r.h, r.x+i, gc.ACS_HLINE)
			renderer.RenderQueue <- render.PositionedChar{
				Rune: '─',
				Cursor: render.Cursor{
					X:     r.x + i,
					Y:     r.y + r.h,
					Style: style,
				},
			}
		}
	}

	// draw corners
	if topBorder && leftBorder {
		// stdscr.MoveAddChar(r.y-1, r.x-1, gc.ACS_ULCORNER)
		renderer.RenderQueue <- render.PositionedChar{
			Rune: '┌',
			Cursor: render.Cursor{
				X:     r.x - 1,
				Y:     r.y - 1,
				Style: style,
			},
		}
	}
	if topBorder && rightBorder {
		// stdscr.MoveAddChar(r.y-1, r.x+r.w, gc.ACS_URCORNER)
		renderer.RenderQueue <- render.PositionedChar{
			Rune: '┐',
			Cursor: render.Cursor{
				X:     r.x + r.w,
				Y:     r.y - 1,
				Style: style,
			},
		}
	}
	if bottomBorder && leftBorder {
		// stdscr.MoveAddChar(r.y+r.h, r.x-1, gc.ACS_LLCORNER)
		renderer.RenderQueue <- render.PositionedChar{
			Rune: '└',
			Cursor: render.Cursor{
				X:     r.x - 1,
				Y:     r.y + r.h,
				Style: style,
			},
		}
	}
	if bottomBorder && rightBorder {
		// stdscr.MoveAddChar(r.y+r.h, r.y+r.w, gc.ACS_LRCORNER)
		renderer.RenderQueue <- render.PositionedChar{
			Rune: '┘',
			Cursor: render.Cursor{
				X:     r.x + r.w,
				Y:     r.y + r.h,
				Style: style,
			},
		}
	}

	// stdscr.Refresh()
	// renderer.Refresh()
}

// getTermSize returns the wusth
func getTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}
