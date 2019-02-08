package main

import (
	"fmt"
	"log"
)

// Direction is the type of Up, Down, Left, and Right
type Direction int

// directions
const (
	_ Direction = iota
	Up
	Down
	Left
	Right
)

// A Path is a series of indicies leading from the root to a Container
type Path []int

func getSelection() Path {
	path := Path{root.selectionIdx}
	selection := root.elements[root.selectionIdx].contents

	for {
		switch val := selection.(type) {
		case *Pane:
			return path
		case *Split:
			path = append(path, val.selectionIdx)
			selection = val.elements[val.selectionIdx].contents
		default:
			panic(fmt.Sprintf("Unexpected type %T", selection))
		}
	}
}

func moveWindow(d Direction) {
	path := getSelection()
	parent, parentPath := path.getParent()

	vert := parent.verticallyStacked

	if (!vert && d == Left) || (vert && d == Up) {
		idx := parent.selectionIdx

		if idx == 0 {
			return
		}

		tmp := parent.elements[idx-1]
		parent.elements[idx-1] = parent.elements[idx]
		parent.elements[idx] = tmp

		parent.selectionIdx--

		// parent.refreshRenderRect()
	} else if (!vert && d == Right) || (vert && d == Down) {
		idx := parent.selectionIdx

		if idx == len(parent.elements)-1 {
			return
		}

		tmp := parent.elements[idx+1]
		parent.elements[idx+1] = parent.elements[idx]
		parent.elements[idx] = tmp

		parent.selectionIdx++

		// parent.refreshRenderRect()
	} else {
		movingVert := d == Up || d == Down

		p := path
		for len(p) > 0 {
			s, _ := p.getParent()
			if s.verticallyStacked == movingVert {
				tmp := parentPath.popContainer((*parent).selectionIdx)

				if d == Left || d == Up {
					s.insertContainer(tmp, s.selectionIdx)
					s.selectionIdx--
				} else {
					s.insertContainer(tmp, s.selectionIdx+1)
					s.selectionIdx++
				}

				// s.refreshRenderRect()
				break
			}
			p = p[:len(p)-1]
		}

		if len(p) == 0 {
			tmp := parentPath.popContainer(parent.selectionIdx)
			tmpRoot := root

			var h int
			if config.statusBar {
				h = termH - 1
			} else {
				h = termH
			}

			root = Split{
				renderRect:        Rect{w: termW, h: h},
				verticallyStacked: movingVert,
				selectionIdx:      0,
				elements: []Node{
					Node{
						size:     1,
						contents: &tmpRoot,
					},
				},
			}

			insertIdx := 0
			if d == Down || d == Right {
				insertIdx = 1
			}
			root.insertContainer(tmp, insertIdx)
			root.selectionIdx = insertIdx

			// fmt.Print("\033[2J")

			// root.refreshRenderRect()
		}
	}
}

func killWindow() {
	parent, parentPath := getSelection().getParent()
	t := parentPath.popContainer(parent.selectionIdx)
	t.(*Pane).kill()
}

func (s *Split) kill() {
	for _, n := range s.elements {
		n.contents.kill()
	}
}

func (p Path) popContainer(idx int) Container {
	s := p.getContainer().(*Split)

	tmp := s.elements[idx]

	s.elements = append(s.elements[:idx], s.elements[idx+1:]...)

	// resize nodes
	scaleFactor := float32(1.0 / (1.0 - tmp.size))
	for i := range s.elements {
		s.elements[i].size *= scaleFactor
	}

	if idx > len(s.elements)-1 {
		s.selectionIdx--
	}

	if len(s.elements) == 1 && len(p) >= 1 {
		switch val := (*s).elements[0].contents.(type) {
		case *Pane:
			parent, _ := p.getParent()
			parent.elements[p[len(p)-1]].contents = val
		case *Split:
			s.verticallyStacked = val.verticallyStacked
			s.elements = val.elements
			s.selectionIdx = val.selectionIdx
		}
	}

	return tmp.contents
}

// stuff like h(h(x), y) -> h(x, y)
func (s *Split) simplify() {
	if len(s.elements) == 1 {
		switch child := (*s).elements[0].contents.(type) {
		case *Split:
			s.verticallyStacked = child.verticallyStacked
			s.elements = child.elements
			s.selectionIdx = child.selectionIdx
		}
	} else {
		newElements := []Node{}
		selectionIdx := (*s).selectionIdx
		for idx, n := range (*s).elements {
			switch child := n.contents.(type) {
			case *Split:
				if child.verticallyStacked == s.verticallyStacked {
					for j := range child.elements {
						child.elements[j].size *= n.size
					}
					newElements = append(newElements, child.elements...)
					if idx == s.selectionIdx {
						selectionIdx += child.selectionIdx
					} else if idx < s.selectionIdx {
						selectionIdx += len(child.elements) - 1
					}
				} else {
					newElements = append(newElements, n)
				}
			case *Pane:
				newElements = append(newElements, n)
			}
		}
		s.elements = newElements
		s.selectionIdx = selectionIdx
	}

	for _, n := range s.elements {
		switch child := n.contents.(type) {
		case *Split:
			child.simplify()
		}
	}
}

func (s *Split) insertContainer(c Container, idx int) {
	newNodeSize := float32(1) / float32(len(s.elements)+1)

	// resize siblings
	scaleFactor := float32(1) - newNodeSize
	for i := range s.elements {
		s.elements[i].size *= scaleFactor
	}

	newNode := Node{
		size:     newNodeSize,
		contents: c,
	}
	s.elements = append(s.elements[:idx], append([]Node{newNode}, s.elements[idx:]...)...)
}

func moveSelection(d Direction) {
	path := getSelection()

	// deselect the old Term
	oldTerm := path.getContainer().(*Pane)
	oldTerm.selected = false
	oldTerm.softRefresh()

	parent, _ := path.getParent()

	vert := parent.verticallyStacked

	if (d == Left && !vert) || (d == Up && vert) {
		parent.selectionIdx--
		if parent.selectionIdx < 0 {
			parent.selectionIdx = 0
		}
	} else if (d == Right && !vert) || (d == Down && vert) {
		parent.selectionIdx++
		if parent.selectionIdx > len(parent.elements)-1 {
			parent.selectionIdx = len(parent.elements) - 1
		}
	} else {
		movingVert := d == Up || d == Down

		p := path
		for len(p) > 0 {
			s, _ := p.getParent()
			if s.verticallyStacked == movingVert {
				if d == Up || d == Left {
					s.selectionIdx--
					if s.selectionIdx < 0 {
						s.selectionIdx = 0
					}
				} else {
					s.selectionIdx++
					if s.selectionIdx > len(s.elements)-1 {
						s.selectionIdx = len(s.elements) - 1
					}
				}
				break
			}
			p = p[:len(p)-1]
		}
	}

	// deselect the old Term
	nowTerm := getSelection().getContainer().(*Pane)
	nowTerm.selected = true
	nowTerm.softRefresh()
}

func newWindow() {
	path := getSelection()

	// deselect the old Term
	oldTerm := path.getContainer().(*Pane)
	oldTerm.selected = false
	// the parent is going to be redrawn so we don't need to redraw the old term right now

	parent, _ := path.getParent()

	size := float32(1) / float32(len(parent.elements)+1)

	// resize siblings
	scaleFactor := float32(1) - size
	for i := range parent.elements {
		parent.elements[i].size *= scaleFactor
	}

	// add new child
	createdTerm := newTerm(true)
	parent.elements = append(parent.elements, Node{
		size:     size,
		contents: createdTerm,
	})

	// update selection to new child
	parent.selectionIdx = len(parent.elements) - 1
	parent.refreshRenderRect()
}

func (p Path) getParent() (*Split, Path) {
	parentPath := p[:len(p)-1]
	return parentPath.getContainer().(*Split), parentPath
}

func (p Path) getContainer() Container {
	if len(p) == 0 {
		return &root
	}

	cur := root.elements[p[0]].contents
	p = p[1:]
	for len(p) > 0 {
		switch val := cur.(type) {
		case *Split:
			cur = val.elements[val.selectionIdx].contents
			p = p[1:]
		default:
			log.Fatal("bad path")
		}
	}

	return cur
}
