package main

import (
	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/render"
)

func drawSelectionBorder(r Rect) {
	leftBorder := r.x > 0
	rightBorder := r.x+r.w+1 < termW
	topBorder := r.y > 0
	bottomBorder := r.y+r.h+1 < termH

	style := render.Style{
		Fg: ecma48.Color{
			ColorMode: ecma48.ColorBit3Normal,
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
