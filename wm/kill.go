package wm

func (u *Universe) Kill() {
	for _, n := range u.workspaces {
		n.contents.Kill()
	}
}

func (s *split) Kill() {
	for _, n := range s.elements {
		n.contents.Kill()
	}
}

func (u *Universe) handleChildDeath(err error) {
	u.onDeath(err) // FIXME: only supports one workspace
}

func (s *workspace) handleChildDeath(err error) {
	s.onDeath(err)
}

func (s *split) handleChildDeath(err error) {
	for idx := len(s.elements) - 1; idx >= 0; idx-- {
		if s.elements[idx].contents.IsDead() {
			s.popElement(idx)
		}
	}
	if len(s.elements) == 0 || err != nil {
		s.Dead = true
		s.onDeath(err)
	} else {
		s.refreshRenderRect(false)
		s.elements[s.selectionIdx].contents.UpdateSelection(s.selected)
	}
}
