package wm

func (u *Universe) CycleSelection(forwards bool) {
	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

	u.workspaces[u.selectionIdx].contents.cycleSelection(forwards)
}

func (s *split) cycleSelection(forwards bool) (bubble bool) {
	if len(s.elements) == 0 {
		return
	}
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		bubble := child.cycleSelection(forwards)
		if bubble {
			if forwards {
				s.selectionIdx = (s.selectionIdx + 1) % len(s.elements)
				switch newChild := s.elements[s.selectionIdx].contents.(type) {
				case Container:
					newChild.selectMin()
				}
			} else {
				s.selectionIdx = (s.selectionIdx - 1 + len(s.elements)) % len(s.elements)
				switch newChild := s.elements[s.selectionIdx].contents.(type) {
				case Container:
					newChild.selectMax()
				}
			}
		}
	case Node:
		if forwards {
			if s.selectionIdx == len(s.elements)-1 {
				return true
			} else {
				s.selectionIdx++
			}

		} else {
			if s.selectionIdx == 0 {
				return true
			} else {
				s.selectionIdx--
			}
		}
	}

	return false
}

func (s *split) selectMin() {
	s.selectionIdx = 0
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		child.selectMin()
	}
}

func (s *split) selectMax() {
	s.selectionIdx = len(s.elements) - 1
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		child.selectMax()
	}
}
