package main

import (
	gc "github.com/rthornton128/goncurses"
)

func drawSelectionBorder(r Rect) {
	leftBorder := r.x > 0
	rightBorder := r.x+r.w+1 < termW
	topBorder := r.y > 0
	bottomBorder := r.y+r.h+1 < termH

	// draw lines
	if leftBorder {
		for i := 0; i <= r.h; i++ {
			stdscr.MoveAddChar(r.y+i, r.x-1, gc.ACS_VLINE)
			// globalCharAggregate <- vterm.Char{
			// 	Rune: '│',
			// 	Cursor: cursor.Cursor{
			// 		X: r.x - 1,
			// 		Y: r.y + i,
			// 		Fg: cursor.Color{
			// 			ColorMode: cursor.ColorBit3Normal,
			// 			Code:      6,
			// 		},
			// 	},
			// }
		}
	}
	if rightBorder {
		for i := 0; i <= r.h; i++ {
			stdscr.MoveAddChar(r.y+i, r.x+r.w, gc.ACS_VLINE)
			// globalCharAggregate <- vterm.Char{
			// 	Rune: '│',
			// 	Cursor: cursor.Cursor{
			// 		X: r.x + r.w,
			// 		Y: r.y + i,
			// 		Fg: cursor.Color{
			// 			ColorMode: cursor.ColorBit3Normal,
			// 			Code:      6,
			// 		},
			// 	},
			// }
		}
	}
	if topBorder {
		for i := 0; i <= r.w; i++ {
			stdscr.MoveAddChar(r.y-1, r.x+i, gc.ACS_HLINE)
			// globalCharAggregate <- vterm.Char{
			// 	Rune: '─',
			// 	Cursor: cursor.Cursor{
			// 		X: r.x + i,
			// 		Y: r.y - 1,
			// 		Fg: cursor.Color{
			// 			ColorMode: cursor.ColorBit3Normal,
			// 			Code:      6,
			// 		},
			// 	},
			// }
		}
	}
	if bottomBorder {
		for i := 0; i <= r.w; i++ {
			stdscr.MoveAddChar(r.y+r.h, r.x+i, gc.ACS_HLINE)
			// globalCharAggregate <- vterm.Char{
			// 	Rune: '─',
			// 	Cursor: cursor.Cursor{
			// 		X: r.x + i,
			// 		Y: r.y + r.h,
			// 		Fg: cursor.Color{
			// 			ColorMode: cursor.ColorBit3Normal,
			// 			Code:      6,
			// 		},
			// 	},
			// }
		}
	}

	// draw corners
	if topBorder && leftBorder {
		stdscr.MoveAddChar(r.y-1, r.x-1, gc.ACS_ULCORNER)
		// globalCharAggregate <- vterm.Char{
		// Rune: '┌',
		// Cursor: cursor.Cursor{
		// 	X: r.x - 1,
		// 	Y: r.y - 1,
		// 	Fg: cursor.Color{
		// 		ColorMode: cursor.ColorBit3Normal,
		// 		Code:      6,
		// 	},
		// },
		// }
	}
	if topBorder && rightBorder {
		stdscr.MoveAddChar(r.y-1, r.x+r.w, gc.ACS_URCORNER)
		// globalCharAggregate <- vterm.Char{
		// 	Rune: '┐',
		// 	Cursor: cursor.Cursor{
		// 		X: r.x + r.w,
		// 		Y: r.y - 1,
		// 		Fg: cursor.Color{
		// 			ColorMode: cursor.ColorBit3Normal,
		// 			Code:      6,
		// 		},
		// 	},
		// }
	}
	if bottomBorder && leftBorder {
		stdscr.MoveAddChar(r.y+r.h, r.x-1, gc.ACS_LLCORNER)
		// globalCharAggregate <- vterm.Char{
		// 	Rune: '└',
		// 	Cursor: cursor.Cursor{
		// 		X: r.x - 1,
		// 		Y: r.y + r.h,
		// 		Fg: cursor.Color{
		// 			ColorMode: cursor.ColorBit3Normal,
		// 			Code:      6,
		// 		},
		// 	},
		// }
	}
	if bottomBorder && rightBorder {
		stdscr.MoveAddChar(r.y+r.h, r.y+r.w, gc.ACS_LRCORNER)
		// globalCharAggregate <- vterm.Char{
		// 	Rune: '┘',
		// 	Cursor: cursor.Cursor{
		// 		X: r.x + r.w,
		// 		Y: r.y + r.h,
		// 		Fg: cursor.Color{
		// 			ColorMode: cursor.ColorBit3Normal,
		// 			Code:      6,
		// 		},
		// 	},
		// }
	}

	stdscr.Refresh()
}
