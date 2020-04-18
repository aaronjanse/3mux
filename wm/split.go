package wm

import (
	"fmt"

	"github.com/aaronjanse/3mux/render"
)

// A Split splits a region of the screen into a areas reserved for multiple child nodes
type split struct {
	verticallyStacked bool
	elements          []SizedNode
	selectionIdx      int
	renderer          *render.Renderer
	renderRect        Rect

	onDeath func(error)
	Dead    bool
	newPane NewPaneFunc
}

func newSplit(renderer *render.Renderer, onDeath func(error), rect Rect, verticallyStacked bool, children []Node, newPane NewPaneFunc) *split {
	s := &split{
		verticallyStacked: verticallyStacked,
		renderer:          renderer,
		onDeath:           onDeath,
		newPane:           newPane,
	}

	if children == nil {
		children = []Node{newPane(renderer)}
	}

	childSize := 1 / float32(len(children))
	s.elements = make([]SizedNode, len(children))
	for i, child := range children {
		child.SetDeathHandler(s.handleChildDeath)
		s.elements[i] = SizedNode{
			size:     childSize,
			contents: child,
		}
	}
	return s
}

func (s *split) IsDead() bool {
	return s.Dead
}

func (s *split) SetDeathHandler(onDeath func(error)) {
	s.onDeath = onDeath
}

func (s *split) Serialize() string {
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
		out += e.contents.Serialize()
	}
	out += ")"

	return out
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (s *split) SetRenderRect(fullscreen bool, x, y, w, h int) {
	s.renderRect = Rect{x, y, w, h}

	// NOTE: should we clear the screen?

	s.refreshRenderRect(fullscreen)
}

func (s *split) GetRenderRect() Rect {
	return s.renderRect
}

// // removeTheDead recursively searches the tree and removes panes with Dead == true.
// // A pane declares itself dead when its shell dies.
// func removeTheDead(path Path) {
// 	s := path.getContainer().(*Split)
// 	for idx := len(s.elements) - 1; idx >= 0; idx-- {
// 		element := s.elements[idx]
// 		switch c := element.contents.(type) {
// 		case *Split:
// 			removeTheDead(append(path, idx))
// 		case *Pane:
// 			if c.Dead {
// 				t := path.popContainer(idx)
// 				t.(*Pane).kill()
// 			}
// 		}
// 	}
// }

// refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// this is for when one or more of a split's children are reshaped
func (s *split) refreshRenderRect(fullscreen bool) {
	x := s.renderRect.X
	y := s.renderRect.Y
	w := s.renderRect.W
	h := s.renderRect.H

	s.redrawLines()

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
			childNode.contents.SetRenderRect(fullscreen, x, y+lastPos+1, w, childArea)
		} else {
			childNode.contents.SetRenderRect(fullscreen, x+lastPos+1, y, childArea, h)
		}
	}
}

func (s *split) redrawLines() {
	x := s.renderRect.X
	y := s.renderRect.Y
	w := s.renderRect.W
	h := s.renderRect.H

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
				s.renderer.HandleCh(render.PositionedChar{
					Rune:   '─',
					Cursor: render.Cursor{X: x + i, Y: y + pos},
				})
			}
		} else {
			for j := 0; j < h; j++ {
				s.renderer.HandleCh(render.PositionedChar{
					Rune:   '│',
					Cursor: render.Cursor{X: x + pos, Y: y + j},
				})
			}
		}
	}
}

func getDividerPositions(area int, contents []SizedNode) []int {
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
