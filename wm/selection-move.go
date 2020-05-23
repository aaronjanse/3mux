package wm

func (u *Universe) MoveSelection(d Direction) {
	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

	u.workspaces[u.selectionIdx].moveSelection(d)
	u.refreshRenderRect()
	u.updateSelection()
}

func (s *workspace) moveSelection(d Direction) {
	if !s.doFullscreen {
		s.contents.moveSelection(d)
	}
}

func (s *split) moveSelection(d Direction) (bubble bool) {
	alignedForwards := (!s.verticallyStacked && d == Right) || (s.verticallyStacked && d == Down)
	alignedBackward := (!s.verticallyStacked && d == Left) || (s.verticallyStacked && d == Up)

	if len(s.elements) == 0 {
		return false
	}
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		bubble := child.moveSelection(d)
		if bubble {
			switch {
			case alignedForwards:
				s.selectionIdx++
				if s.selectionIdx > len(s.elements)-1 {
					s.selectionIdx = len(s.elements) - 1
					return true
				}
			case alignedBackward:
				s.selectionIdx--
				if s.selectionIdx < 0 {
					s.selectionIdx = 0
					return true
				}
			default:
				return true
			}
		}
	case Node:
		if alignedBackward {
			if s.selectionIdx <= 0 {
				return true
			} else {
				s.selectionIdx--
			}
		} else if alignedForwards {
			if s.selectionIdx >= len(s.elements)-1 {
				return true
			} else {
				s.selectionIdx++
			}
		} else {
			return true
		}
	default:
		panic("should never happen")
	}

	return false
}
