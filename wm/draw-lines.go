package wm

import (
	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/render"
)

func (s *split) drawSelectionBorder() {
	if !s.selected {
		return
	}

	r := s.elements[s.selectionIdx].contents.GetRenderRect()

	style := render.Style{
		Fg: ecma48.Color{
			ColorMode: ecma48.ColorBit3Normal,
			Code:      6,
		},
	}

	for i := 0; i <= r.H; i++ {
		ch := render.PositionedChar{
			Rune: '│',
			Cursor: render.Cursor{
				X:     r.X - 1,
				Y:     r.Y + i,
				Style: style,
			},
		}

		s.renderer.HandleCh(ch)
	}
	for i := 0; i <= r.H; i++ {
		ch := render.PositionedChar{
			Rune: '│',
			Cursor: render.Cursor{
				X:     r.X + r.W,
				Y:     r.Y + i,
				Style: style,
			},
		}

		s.renderer.HandleCh(ch)
	}
	for i := 0; i <= r.W; i++ {
		ch := render.PositionedChar{
			Rune: '─',
			Cursor: render.Cursor{
				X:     r.X + i,
				Y:     r.Y - 1,
				Style: style,
			},
		}

		s.renderer.HandleCh(ch)
	}
	for i := 0; i <= r.W; i++ {
		ch := render.PositionedChar{
			Rune: '─',
			Cursor: render.Cursor{
				X:     r.X + i,
				Y:     r.Y + r.H,
				Style: style,
			},
		}

		s.renderer.HandleCh(ch)
	}

	ch := render.PositionedChar{
		Rune: '┌',
		Cursor: render.Cursor{
			X:     r.X - 1,
			Y:     r.Y - 1,
			Style: style,
		},
	}

	s.renderer.HandleCh(ch)
	ch = render.PositionedChar{
		Rune: '┐',
		Cursor: render.Cursor{
			X:     r.X + r.W,
			Y:     r.Y - 1,
			Style: style,
		},
	}

	s.renderer.HandleCh(ch)
	ch = render.PositionedChar{
		Rune: '└',
		Cursor: render.Cursor{
			X:     r.X - 1,
			Y:     r.Y + r.H,
			Style: style,
		},
	}

	s.renderer.HandleCh(ch)
	ch = render.PositionedChar{
		Rune: '┘',
		Cursor: render.Cursor{
			X:     r.X + r.W,
			Y:     r.Y + r.H,
			Style: style,
		},
	}

	s.renderer.HandleCh(ch)
}
