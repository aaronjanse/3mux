package main

// func drawSelectionBorder(r Rect) {
// 	leftBorder := r.x > 0
// 	rightBorder := r.x+r.w+1 < termW
// 	topBorder := r.y > 0
// 	bottomBorder := r.y+r.h+1 < termH

// 	// draw lines
// 	if leftBorder {
// 		for i := 0; i <= r.h; i++ {
// 			globalCharAggregate <- vterm.Char{
// 				Rune: '│',
// 				Cursor: cursor.Cursor{
// 					X: r.x - 1,
// 					Y: r.y + i,
// 					Fg: cursor.Color{
// 						ColorMode: cursor.ColorBit3Normal,
// 						Code:      6,
// 					},
// 				},
// 			}
// 		}
// 	}
// 	if rightBorder {
// 		for i := 0; i <= r.h; i++ {
// 			globalCharAggregate <- vterm.Char{
// 				Rune: '│',
// 				Cursor: cursor.Cursor{
// 					X: r.x + r.w,
// 					Y: r.y + i,
// 					Fg: cursor.Color{
// 						ColorMode: cursor.ColorBit3Normal,
// 						Code:      6,
// 					},
// 				},
// 			}
// 		}
// 	}
// 	if topBorder {
// 		for i := 0; i <= r.w; i++ {
// 			globalCharAggregate <- vterm.Char{
// 				Rune: '─',
// 				Cursor: cursor.Cursor{
// 					X: r.x + i,
// 					Y: r.y - 1,
// 					Fg: cursor.Color{
// 						ColorMode: cursor.ColorBit3Normal,
// 						Code:      6,
// 					},
// 				},
// 			}
// 		}
// 	}
// 	if bottomBorder {
// 		for i := 0; i <= r.w; i++ {
// 			globalCharAggregate <- vterm.Char{
// 				Rune: '─',
// 				Cursor: cursor.Cursor{
// 					X: r.x + i,
// 					Y: r.y + r.h,
// 					Fg: cursor.Color{
// 						ColorMode: cursor.ColorBit3Normal,
// 						Code:      6,
// 					},
// 				},
// 			}
// 		}
// 	}

// 	// draw corners
// 	if topBorder && leftBorder {
// 		globalCharAggregate <- vterm.Char{
// 			Rune: '┌',
// 			Cursor: cursor.Cursor{
// 				X: r.x - 1,
// 				Y: r.y - 1,
// 				Fg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Normal,
// 					Code:      6,
// 				},
// 			},
// 		}
// 	}
// 	if topBorder && rightBorder {
// 		globalCharAggregate <- vterm.Char{
// 			Rune: '┐',
// 			Cursor: cursor.Cursor{
// 				X: r.x + r.w,
// 				Y: r.y - 1,
// 				Fg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Normal,
// 					Code:      6,
// 				},
// 			},
// 		}
// 	}
// 	if bottomBorder && leftBorder {
// 		globalCharAggregate <- vterm.Char{
// 			Rune: '└',
// 			Cursor: cursor.Cursor{
// 				X: r.x - 1,
// 				Y: r.y + r.h,
// 				Fg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Normal,
// 					Code:      6,
// 				},
// 			},
// 		}
// 	}
// 	if bottomBorder && rightBorder {
// 		globalCharAggregate <- vterm.Char{
// 			Rune: '┘',
// 			Cursor: cursor.Cursor{
// 				X: r.x + r.w,
// 				Y: r.y + r.h,
// 				Fg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Normal,
// 					Code:      6,
// 				},
// 			},
// 		}
// 	}
// }
