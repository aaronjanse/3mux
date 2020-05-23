package wm

import (
	"errors"
)

func (u *Universe) MoveWindow(dir Direction) error {
	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

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
	bubble, _, p := s.contents.moveWindow(dir)
	if bubble {
		if dir == Up || dir == Left {
			s.contents = newSplit(
				s.renderer, s.contents.u, s.handleChildDeath, s.renderRect,
				(dir == Up), 0, []Node{p, s.contents}, s.newPane,
			)
		} else {
			s.contents = newSplit(
				s.renderer, s.contents.u, s.handleChildDeath, s.renderRect,
				(dir == Down), 1, []Node{s.contents, p}, s.newPane,
			)
		}
	}
	return nil
}

func (s *split) moveWindow(d Direction) (bubble bool, superBubble bool, p Node) {
	alignedForwards := (!s.verticallyStacked && d == Right) || (s.verticallyStacked && d == Down)
	alignedBackward := (!s.verticallyStacked && d == Left) || (s.verticallyStacked && d == Up)

	if len(s.elements) == 0 {
		return
	}
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		bubble, superBubble, p := child.moveWindow(d)
		if superBubble {
			return true, false, p
		}
		if bubble {
			var idx int
			switch {
			case alignedForwards:
				idx = s.selectionIdx + 1
			case alignedBackward:
				idx = s.selectionIdx
			default: // VSplit[0](HSplit[0](Term[0,0 126x8]*), Term[0,9 126x8])
				return true, false, p
			}

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

			s.elements = append(
				s.elements[:idx], append(
					[]SizedNode{newNode},
					s.elements[idx:]...)...)
		}
	case Node:
		if alignedBackward {
			if s.selectionIdx == 0 {
				s.popElement(s.selectionIdx)
				return true, false, child
			}

			tmp := s.elements[s.selectionIdx-1]
			s.elements[s.selectionIdx-1] = s.elements[s.selectionIdx]
			s.elements[s.selectionIdx] = tmp

			s.selectionIdx--
		} else if alignedForwards {
			if s.selectionIdx == len(s.elements)-1 {
				s.popElement(s.selectionIdx)
				return true, false, child
			}

			tmp := s.elements[s.selectionIdx+1]
			s.elements[s.selectionIdx+1] = s.elements[s.selectionIdx]
			s.elements[s.selectionIdx] = tmp

			s.selectionIdx++
		} else {
			switch len(s.elements) {
			case 0:
				panic("cannot move without elements")
			case 2:
				s.verticallyStacked = !s.verticallyStacked
			default:
				s.popElement(s.selectionIdx)
				return true, len(s.elements) == 0, child
			}
		}
	}

	return false, false, nil
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
