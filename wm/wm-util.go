package wm

import (
	"fmt"
	"log"

	"github.com/aaronjanse/3mux/pane"
)

// import (
// 	"fmt"
// )

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

func (u *Universe) getSelection() Path {
	path := Path{u.selectionIdx}
	wsSplit := u.workspaces[root.selectionIdx].contents

	path = append(path, wsSplit.selectionIdx)
	selection := wsSplit.elements[wsSplit.selectionIdx].contents

	for {
		switch val := selection.(type) {
		case *pane.Pane:
			return path
		case *Split:
			path = append(path, val.selectionIdx)
			selection = val.elements[val.selectionIdx].contents
		default:
			panic(fmt.Sprintf("Unexpected type %T", selection))
		}
	}
}

// func setSelection(path Path) {
// 	root.selectionIdx = path[0]

// 	split := root.workspaces[root.selectionIdx].contents

// 	for i := range path {
// 		if i == 0 {
// 			continue
// 		} else if i > 1 {
// 			split = split.elements[split.selectionIdx].contents.(*Split)
// 		}

// 		split.selectionIdx = path[i]
// 	}
// }

func (p Path) getParent(root *Universe) (*Split, Path) {
	parentPath := p[:len(p)-1]
	return parentPath.getContainer(root).(*Split), parentPath
}

func (p Path) getContainer(root *Universe) Container {
	if len(p) == 0 {
		return root
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
			panic("bad path")
		}
	}

	return cur
}

// func getPanes() []*Pane {
// 	return getPanesOfSplit(root.workspaces[root.selectionIdx].contents)
// }

// func getPanesOfSplit(s *Split) []*Pane {
// 	panes := []*Pane{}
// 	for _, e := range s.elements {
// 		switch c := e.contents.(type) {
// 		case *Split:
// 			panes = append(panes, getPanesOfSplit(c)...)
// 		case *Pane:
// 			panes = append(panes, c)
// 		}
// 	}

// 	return panes
// }

func (p Path) popContainer(u *Universe, idx int) Container {
	s := p.getContainer(u).(*Split)

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
		case *pane.Pane:
			parent, _ := p.getParent(u)
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

func removeTheDead() {
	implRemoveTheDead([]int{0})
}

// removeTheDead recursively searches the tree and removes panes with Dead == true.
// A pane declares itself dead when its shell dies.
func implRemoveTheDead(path Path) {
	log.Println("REMOVING THE DEAD")
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
}
