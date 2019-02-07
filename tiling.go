package main

import "fmt"

// A Node represents a single pane of a split, having a size (relative to a total 1.0) and renderable contents
type Node struct {
	size     float32
	contents Container
}

// A Container is a renderable, debuggable, and killable unit of the window tree
type Container interface {
	setRenderRect(x, y, w, h int)
	serialize() string
	kill()
}

// A Split splits a region of the screen into a areas reserved for multiple child nodes
type Split struct {
	elements          []Node
	selectionIdx      int
	verticallyStacked bool

	renderRect Rect
}

func (t *Pane) serialize() string {
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

var root = Split{
	verticallyStacked: false,
	selectionIdx:      0,
	elements: []Node{
		Node{
			size:     1,
			contents: newTerm(true),
		},
	}}
