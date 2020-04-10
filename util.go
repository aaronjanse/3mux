package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aaronjanse/3mux/render"
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
		for i := 0; i < r.h; i++ {
			ch := render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.x - 1,
					Y:     r.y + i,
					Style: style,
				},
			}

			renderer.HandleCh(ch)
		}
	}
	if rightBorder {
		for i := 0; i < r.h; i++ {
			ch := render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.x + r.w,
					Y:     r.y + i,
					Style: style,
				},
			}

			renderer.HandleCh(ch)
		}
	}
	if topBorder {
		for i := 0; i <= r.w; i++ {
			ch := render.PositionedChar{
				Rune: '─',
				Cursor: render.Cursor{
					X:     r.x + i,
					Y:     r.y - 1,
					Style: style,
				},
			}

			renderer.HandleCh(ch)
		}
	}
	if bottomBorder {
		for i := 0; i <= r.w; i++ {
			ch := render.PositionedChar{
				Rune: '─',
				Cursor: render.Cursor{
					X:     r.x + i,
					Y:     r.y + r.h,
					Style: style,
				},
			}

			renderer.HandleCh(ch)
		}
	}

	// draw corners
	if topBorder && leftBorder {
		ch := render.PositionedChar{
			Rune: '┌',
			Cursor: render.Cursor{
				X:     r.x - 1,
				Y:     r.y - 1,
				Style: style,
			},
		}

		renderer.HandleCh(ch)
	}
	if topBorder && rightBorder {
		ch := render.PositionedChar{
			Rune: '┐',
			Cursor: render.Cursor{
				X:     r.x + r.w,
				Y:     r.y - 1,
				Style: style,
			},
		}

		renderer.HandleCh(ch)
	}
	if bottomBorder && leftBorder {
		ch := render.PositionedChar{
			Rune: '└',
			Cursor: render.Cursor{
				X:     r.x - 1,
				Y:     r.y + r.h,
				Style: style,
			},
		}

		renderer.HandleCh(ch)
	}
	if bottomBorder && rightBorder {
		ch := render.PositionedChar{
			Rune: '┘',
			Cursor: render.Cursor{
				X:     r.x + r.w,
				Y:     r.y + r.h,
				Style: style,
			},
		}

		renderer.HandleCh(ch)
	}
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
