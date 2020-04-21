package wm

import (
	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/render"
)

func (u *Universe) getSelectedNode() Node {
	return u.workspaces[u.selectionIdx].getSelectedNode()
}

func (s *workspace) getSelectedNode() Node {
	return s.contents.getSelectedNode()
}

func (s *split) getSelectedNode() Node {
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		return child.getSelectedNode()
	case Node:
		return child
	}
	panic("should never happen")
}

func (u *Universe) drawSelectionBorder() {
	maxH := u.workspaces[u.selectionIdx].contents.GetRenderRect().H

	r := u.getSelectedNode().GetRenderRect()

	style := render.Style{
		Fg: ecma48.Color{
			ColorMode: ecma48.ColorBit3Normal,
			Code:      6,
		},
	}

	for i := 0; i <= r.H; i++ {
		if r.Y+i < maxH {
			ch := render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.X - 1,
					Y:     r.Y + i,
					Style: style,
				},
			}
			u.renderer.HandleCh(ch)
		}
	}
	for i := 0; i <= r.H; i++ {
		if r.Y+i < maxH {
			ch := render.PositionedChar{
				Rune: '│',
				Cursor: render.Cursor{
					X:     r.X + r.W,
					Y:     r.Y + i,
					Style: style,
				},
			}
			u.renderer.HandleCh(ch)
		}
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
		u.renderer.HandleCh(ch)
	}

	if r.Y+r.H < maxH {
		for i := 0; i <= r.W; i++ {
			ch := render.PositionedChar{
				Rune: '─',
				Cursor: render.Cursor{
					X:     r.X + i,
					Y:     r.Y + r.H,
					Style: style,
				},
			}

			u.renderer.HandleCh(ch)
		}
	}

	ch := render.PositionedChar{
		Rune: '┌',
		Cursor: render.Cursor{
			X:     r.X - 1,
			Y:     r.Y - 1,
			Style: style,
		},
	}
	u.renderer.HandleCh(ch)

	ch = render.PositionedChar{
		Rune: '┐',
		Cursor: render.Cursor{
			X:     r.X + r.W,
			Y:     r.Y - 1,
			Style: style,
		},
	}
	u.renderer.HandleCh(ch)

	if r.Y+r.H < maxH {
		ch = render.PositionedChar{
			Rune: '└',
			Cursor: render.Cursor{
				X:     r.X - 1,
				Y:     r.Y + r.H,
				Style: style,
			},
		}
		u.renderer.HandleCh(ch)

		ch = render.PositionedChar{
			Rune: '┘',
			Cursor: render.Cursor{
				X:     r.X + r.W,
				Y:     r.Y + r.H,
				Style: style,
			},
		}
		u.renderer.HandleCh(ch)
	}
}
