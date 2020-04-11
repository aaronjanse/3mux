package main

import (
	"fmt"
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
	wsSplit := root.workspaces[root.selectionIdx].contents

	path = append(path, wsSplit.selectionIdx)
	selection := wsSplit.elements[wsSplit.selectionIdx].contents

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

func setSelection(path Path) {
	root.selectionIdx = path[0]

	split := root.workspaces[root.selectionIdx].contents

	for i := range path {
		if i == 0 {
			continue
		} else if i > 1 {
			split = split.elements[split.selectionIdx].contents.(*Split)
		}

		split.selectionIdx = path[i]
	}
}

func (p Path) getParent() (*Split, Path) {
	parentPath := p[:len(p)-1]
	return parentPath.getContainer().(*Split), parentPath
}

func (p Path) getContainer() Container {
	if len(p) == 0 {
		return &root
	}

	wsSplit := root.workspaces[p[0]].contents

	if len(p) == 1 {
		return wsSplit
	}

	cur := wsSplit.elements[p[1]].contents
	p = p[2:]
	for len(p) > 0 {
		switch val := cur.(type) {
		case *Split:
			cur = val.elements[val.selectionIdx].contents
			p = p[1:]
		default:
			fatalShutdownNow("bad path")
		}
	}

	return cur
}

func getPanes() []*Pane {
	return getPanesOfSplit(root.workspaces[root.selectionIdx].contents)
}

func getPanesOfSplit(s *Split) []*Pane {
	panes := []*Pane{}
	for _, e := range s.elements {
		switch c := e.contents.(type) {
		case *Split:
			panes = append(panes, getPanesOfSplit(c)...)
		case *Pane:
			panes = append(panes, c)
		}
	}

	return panes
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

	if len(s.elements) == 1 && len(p) > 1 {
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
