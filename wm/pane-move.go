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

	var totalSize float32
	for _, e := range s.elements {
		totalSize += e.size
	}
	scale := 1 / totalSize
	for i := range s.elements {
		s.elements[i].size *= scale
	}
}
