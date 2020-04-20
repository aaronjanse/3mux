package wm

func (u *Universe) KillPane() {
	allDead := u.workspaces[u.selectionIdx].killPane()
	if !allDead {
		u.refreshRenderRect()
		u.updateSelection()
	}
}

func (s *workspace) killPane() (dead bool) {
	s.setFullscreen(false)
	return s.contents.killPane()
}

func (s *split) killPane() (dead bool) {
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		child.killPane()
	case Node:
		child.Kill()
	}

	for i := len(s.elements) - 1; i >= 0; i-- {
		if s.elements[i].contents.IsDead() {
			s.popElement(i)
		}
	}

	if s.selectionIdx >= len(s.elements) {
		s.selectionIdx = len(s.elements) - 1
	}

	if len(s.elements) == 0 {
		s.Dead = true
		return true
	}
	return false
}
