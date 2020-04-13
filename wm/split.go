package wm

import (
	"fmt"

	"github.com/aaronjanse/3mux/pane"
	gc "github.com/rthornton128/goncurses"
)

// A Split splits a region of the screen into a areas reserved for multiple child nodes
type Split struct {
	elements          []Node
	selectionIdx      int
	verticallyStacked bool
	stdscr            *gc.Window

	renderRect Rect
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (s *Split) Reshape(x, y, w, h int) {
	s.renderRect = Rect{x, y, w, h}

	// NOTE: should we clear the screen?

	// s.refreshRenderRect()
}

// func (s *Split) getRenderRect() Rect {
// 	return s.renderRect
// }

func (s *Split) Kill() {
	for _, n := range s.elements {
		n.contents.Kill()
	}
}

func (s *Split) AddPane() {
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case *Split:
		child.AddPane()
		return
	}

	// we assume we're the last split in the selection chain

	if len(s.elements) > 8 {
		return
	}

	size := float32(1) / float32(len(s.elements)+1)

	// resize siblings
	scaleFactor := float32(1) - size
	for i := range s.elements {
		s.elements[i].size *= scaleFactor
	}

	var length float32
	if s.verticallyStacked {
		length = float32(s.renderRect.h)
	} else {
		length = float32(s.renderRect.w)
	}

	var p *pane.Pane
	var err error
	if s.verticallyStacked {
		p, err = pane.NewPane(
			s.renderRect.x,
			s.renderRect.y+int(size*length)+1,
			s.renderRect.w,
			int(float32(s.renderRect.h)*size)-1,
			removeTheDead,
		)
	} else {
		p, err = pane.NewPane(
			s.renderRect.x+int(size*length)+1,
			s.renderRect.y,
			int(float32(s.renderRect.w)*size)-1,
			s.renderRect.h,
			removeTheDead,
		)
	}
	if err != nil {
		panic(err)
	}

	s.elements = append(s.elements, Node{
		size:     size,
		contents: p,
	})

	// update selection to new child
	s.selectionIdx = len(s.elements) - 1

	s.refreshChildShapes()
	s.drawLines()
}

// func (s *Split) setPause(pause bool) {
// 	for _, e := range s.elements {
// 		e.contents.setPause(pause)
// 	}
// }

func (s *Split) refreshChildShapes() {
	x := s.renderRect.x
	y := s.renderRect.y
	w := s.renderRect.w
	h := s.renderRect.h

	var area int
	if s.verticallyStacked {
		area = h
	} else {
		area = w
	}
	dividers := getDividerPositions(area, s.elements)
	if len(s.elements) == 1 {
		dividers = []int{area}
	}
	for idx, pos := range dividers {
		lastPos := -1
		if idx > 0 {
			lastPos = dividers[idx-1]
		}

		childArea := pos - lastPos - 1
		if idx == len(dividers)-1 && idx != 0 {
			childArea = area - lastPos - 1
		}

		childNode := s.elements[idx]

		if s.verticallyStacked {
			childNode.contents.Reshape(x, y+lastPos+1, w, childArea)
		} else {
			childNode.contents.Reshape(x+lastPos+1, y, childArea, h)
		}
	}

	s.drawLines()
}

func (s *Split) drawLines() {
	x := s.renderRect.x
	y := s.renderRect.y
	w := s.renderRect.w
	h := s.renderRect.h

	var area int
	if s.verticallyStacked {
		area = h
	} else {
		area = w
	}
	dividers := getDividerPositions(area, s.elements)
	for idx, pos := range dividers {
		if idx == len(dividers)-1 {
			break
		}

		if s.verticallyStacked {
			for i := 0; i < w; i++ {
				s.stdscr.Move(y+pos, x+i)
				s.stdscr.Print("─")
				// s.stdscr.MoveAddChar(y+pos, x+i, '-')
				// renderer.HandleCh(render.PositionedChar{
				// 	Rune:   '─',
				// 	Cursor: render.Cursor{X: x + i, Y: y + pos},
				// })
			}
		} else {
			for j := 0; j < h; j++ {
				// renderer.HandleCh(render.PositionedChar{
				// 	Rune:   '│',
				// 	Cursor: render.Cursor{X: x + pos, Y: y + j},
				// })
				s.stdscr.Move(y+j, x+pos)
				s.stdscr.Print("│")
				// s.stdscr.MoveAddChar(y+j, x+pos, '|')
			}
		}
	}
}

func getDividerPositions(area int, contents []Node) []int {
	var dividerPositions []int
	for idx, node := range contents {
		var lastPos int
		if idx == 0 {
			lastPos = 0
		} else {
			lastPos = dividerPositions[idx-1]
		}
		pos := lastPos + int(node.size*float32(area))
		dividerPositions = append(dividerPositions, pos)
	}
	return dividerPositions
}

func (s *Split) Serialize() string {
	var out string
	if s.verticallyStacked {
		out = "VSplit"
	} else {
		out = "HSplit"
	}

	out += fmt.Sprintf("[%d]", s.selectionIdx)

	out += "("
	for i, e := range s.elements {
		if i != 0 {
			out += ", "
		}
		// out += fmt.Sprintf("{%f}", e.size)
		out += e.contents.Serialize()
	}
	out += ")"

	return out
}
