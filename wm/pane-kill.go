package wm

func (u *Universe) KillPane() {
	u.workspaces[u.selectionIdx].killPane()
	u.refreshRenderRect()
	u.updateSelection()
}

func (s *workspace) killPane() {
	s.setFullscreen(false)
	s.contents.killPane()
}

func (s *split) killPane() {
	switch child := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		child.killPane()
	case Node:
		child.Kill()
	}

	allDead := true
	for _, n := range s.elements {
		allDead = allDead && n.contents.IsDead()
	}

	if allDead {
		s.Dead = true
		s.onDeath(nil)
	}
}
