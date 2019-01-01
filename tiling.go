package main

import "fmt"

type Node struct {
	size     float32
	contents Container
}

type Container interface {
	setRenderRect(x, y, w, h int)
	serialize() string
}

// A Split splits a region of the screen into a areas reserved for multiple child nodes
type Split struct {
	elements          []Node
	selectionIdx      int
	verticallyStacked bool

	renderRect Rect
}

func (t *Term) serialize() string {
	return fmt.Sprintf("Term")
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

// var root = Split{
// 	verticallyStacked: false,
// 	selectionIdx:      0,
// 	elements: []Node{
// 		Node{
// 			size:     0.4,
// 			contents: &Term{id: 0},
// 		},
// 		Node{
// 			size: 0.6,
// 			contents: &Split{
// 				verticallyStacked: true,
// 				selectionIdx:      1,
// 				elements: []Node{
// 					Node{
// 						size:     0.7,
// 						contents: &Term{id: 1},
// 					},
// 					Node{
// 						size:     0.3,
// 						contents: &Term{id: 2},
// 					},
// 				}},
// 		},
// 	}}

var root = Split{
	verticallyStacked: false,
	selectionIdx:      0,
	elements: []Node{
		Node{
			size:     1,
			contents: newTerm(),
		},
	}}
