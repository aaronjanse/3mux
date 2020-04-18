package wm

import (
	"errors"
)

func (u *Universe) MoveWindow(dir Direction) error {
	err := u.workspaces[u.selectionIdx].moveWindow(dir)
	if err != nil {
		return err
	}
	u.simplify()
	u.refreshRenderRect()
	return nil
}

func (s *workspace) moveWindow(dir Direction) error {
	if s.doFullscreen {
		return errors.New("cannot move window while one is fullscreen")
	}
	s.contents.moveWindow(dir)
	return nil
}

func (s *split) moveWindow(d Direction) (bubble bool, p Node) {
	alignedForwards := (!s.verticallyStacked && d == Right) || (s.verticallyStacked && d == Down)
	alignedBackward := (!s.verticallyStacked && d == Left) || (s.verticallyStacked && d == Up)

	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		bubble, p := child.moveWindow(d)
		if bubble {
			p.SetDeathHandler(s.handleChildDeath)

			newNodeSize := float32(1) / float32(len(s.elements)+1)

			// resize siblings
			scaleFactor := float32(1) - newNodeSize
			for i := range s.elements {
				s.elements[i].size *= scaleFactor
			}

			newNode := SizedNode{
				size:     newNodeSize,
				contents: p,
			}

			var idx int
			switch {
			case alignedForwards:
				idx = s.selectionIdx + 1
			case alignedBackward:
				idx = s.selectionIdx
			default:
				panic("should never happen")
			}

			s.elements = append(
				s.elements[:idx], append(
					[]SizedNode{newNode},
					s.elements[idx:]...)...)
		}
	case Node:
		if alignedBackward {
			if s.selectionIdx == 0 {
				s.popElement(s.selectionIdx)
				return true, child
			} else {
				tmp := s.elements[s.selectionIdx-1]
				s.elements[s.selectionIdx-1] = s.elements[s.selectionIdx]
				s.elements[s.selectionIdx] = tmp

				s.selectionIdx--
			}
		} else if alignedForwards {
			if s.selectionIdx == len(s.elements)-1 {
				s.popElement(s.selectionIdx)
				return true, child
			} else {
				tmp := s.elements[s.selectionIdx+1]
				s.elements[s.selectionIdx+1] = s.elements[s.selectionIdx]
				s.elements[s.selectionIdx] = tmp

				s.selectionIdx++
			}
		} else {
			switch len(s.elements) {
			case 0:
				panic("cannot move without elements")
			case 2:
				s.verticallyStacked = !s.verticallyStacked
			default:
				s.popElement(s.selectionIdx)
				return true, child
			}
		}
	}

	return false, nil
}

func (s *split) popElement(idx int) {
	s.elements = append(s.elements[:idx], s.elements[idx+1:]...)
	if s.selectionIdx > len(s.elements)-1 {
		s.selectionIdx = len(s.elements) - 1
	}
}

// func (s *Split) moveWindowXX(d Direction) {
// 	path := getSelection()
// 	parent, parentPath := path.getParent()

// 	vert := parent.verticallyStacked

// 	{
// 		movingVert := d == Up || d == Down

// 		p := path
// 		for len(p) > 1 {
// 			s, _ := p.getParent()
// 			if s.verticallyStacked == movingVert {
// 				tmp := parentPath.popContainer(parent.selectionIdx)

// 				if d == Left || d == Up {
// 					s.insertContainer(tmp, s.selectionIdx)
// 				} else {
// 					s.insertContainer(tmp, s.selectionIdx+1)
// 					s.selectionIdx++
// 				}

// 				// root.refreshRenderRect()
// 				break
// 			}
// 			p = p[:len(p)-1]
// 		}

// 		// if len(p) == 1 && len(parent.elements) > 1 {
// 		// 	tmp := parentPath.popContainer(parent.selectionIdx)
// 		// 	tmpRoot := root.workspaces[root.selectionIdx].contents

// 		// 	var h int
// 		// 	if config.statusBar {
// 		// 		h = termH - 1
// 		// 	} else {
// 		// 		h = termH
// 		// 	}

// 		// 	root.workspaces[root.selectionIdx].contents = NewSplit(
// 		// 		renderer, movingVert, Rect{x: 0, y: 0, w: termW, h: h},
// 		// 		[]Container{tmpRoot},
// 		// 	)

// 		// 	insertIdx := 0
// 		// 	if d == Down || d == Right {
// 		// 		insertIdx = 1
// 		// 	}
// 		// 	root.workspaces[root.selectionIdx].contents.insertContainer(tmp, insertIdx)
// 		// 	root.workspaces[root.selectionIdx].contents.selectionIdx = insertIdx

// 		// 	// root.refreshRenderRect()
// 		// }
// 	}

// 	// select the new Term
// 	newTerm := getSelection().getContainer().(*Pane)
// 	newTerm.selected = true
// 	newTerm.softRefresh()
// 	newTerm.vterm.RefreshCursor()

// 	root.simplify()
// 	root.refreshRenderRect()
// }
