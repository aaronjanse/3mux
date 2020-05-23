package wm

func (u *Universe) ResizePane(d Direction) {
	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

	u.workspaces[u.selectionIdx].resizePane(d)
	u.refreshRenderRect()
}

func (s *workspace) resizePane(d Direction) {
	s.contents.resizePane(d)
}

func (s *split) resizePane(d Direction) (bubble bool) {
	if s.selectionIdx >= len(s.elements) {
		s.selectionIdx = len(s.elements) - 1
	}
	if len(s.elements) == 0 {
		return false
	}
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		bubble := child.resizePane(d)
		if bubble {
			s.resizeSelectedChild(d)
		}
	case Node:
		if s.verticallyStacked {
			if d == Left || d == Right {
				return true
			}
		} else {
			if d == Up || d == Down {
				return true
			}
		}

		s.resizeSelectedChild(d)
	}

	return false
}

func (s *split) resizeSelectedChild(d Direction) {
	selectedSize := s.elements[s.selectionIdx].size

	var delta float32
	if s.verticallyStacked {
		if d == Up {
			delta = -0.1
		} else {
			delta = +0.1
		}
	} else {
		if d == Left {
			delta = -0.1
		} else {
			delta = +0.1
		}
	}

	var length float32
	if s.verticallyStacked {
		length = float32(s.renderRect.H)
	} else {
		length = float32(s.renderRect.W)
	}

	var min float32 = 1
	if (selectedSize+delta)*length <= min {
		return
	}

	scale := (float32(1) - selectedSize - delta) / (float32(1) - selectedSize)

	for i := range s.elements {
		if i != s.selectionIdx {
			elem := s.elements[i]
			newSize := elem.size * scale

			var wh float32
			if s.verticallyStacked {
				wh = float32(elem.contents.GetRenderRect().H)
			} else {
				wh = float32(elem.contents.GetRenderRect().W)
			}

			if newSize*wh <= min {
				return
			}
		}
	}

	s.elements[s.selectionIdx].size += delta
	for i := range s.elements {
		if i != s.selectionIdx {
			s.elements[i].size *= scale
		}
	}
}
