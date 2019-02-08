package main

import (
	"fmt"
)

// A Split splits a region of the screen into a areas reserved for multiple child nodes
type Split struct {
	elements          []Node
	selectionIdx      int
	verticallyStacked bool

	renderRect Rect
}

func (s *Split) serialize() string {
	var out string
	if s.verticallyStacked {
		out = "VSplit"
	} else {
		out = "HSplit"
	}

	out += fmt.Sprintf("[%d]", s.selectionIdx)

	out += "("
	for _, e := range s.elements {
		out += e.contents.serialize() + ", "
	}
	out += ")"

	return out
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (s *Split) setRenderRect(x, y, w, h int) {
	s.renderRect = Rect{x, y, w, h}

	// // clear the relevant area of the screen
	// for j := 0; j < h; j++ {
	// 	for i := 0; i < w; i++ {
	// 		globalCharAggregate <- vterm.Char{
	// 			Rune:   '~',
	// 			Cursor: cursor.Cursor{X: x + i, Y: y + j},
	// 		}
	// 	}
	// }

	s.refreshRenderRect()
}

// refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// this is for when one or more of a split's children are reshaped
func (s *Split) refreshRenderRect() {
	x := s.renderRect.x
	y := s.renderRect.y
	w := s.renderRect.w
	h := s.renderRect.h

	// s.redrawLines()

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
			childNode.contents.setRenderRect(x, y+lastPos+1, w, childArea)
		} else {
			childNode.contents.setRenderRect(x+lastPos+1, y, childArea, h)
		}
	}
}

// func (s *Split) redrawLines() {
// 	x := s.renderRect.x
// 	y := s.renderRect.y
// 	w := s.renderRect.w
// 	h := s.renderRect.h

// 	var area int
// 	if s.verticallyStacked {
// 		area = h
// 	} else {
// 		area = w
// 	}
// 	dividers := getDividerPositions(area, s.elements)
// 	for idx, pos := range dividers {
// 		if idx == len(dividers)-1 {
// 			break
// 		}

// 		// if s.verticallyStacked {
// 		// 	for i := 0; i < w; i++ {
// 		// 		globalCharAggregate <- vterm.Char{
// 		// 			Rune:   '─',
// 		// 			Cursor: cursor.Cursor{X: x + i, Y: y + pos},
// 		// 		}
// 		// 	}
// 		// } else {
// 		// 	for j := 0; j < h; j++ {
// 		// 		globalCharAggregate <- vterm.Char{
// 		// 			Rune:   '│',
// 		// 			Cursor: cursor.Cursor{X: x + pos, Y: y + j},
// 		// 		}
// 		// 	}
// 		// }
// 	}
// }

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
