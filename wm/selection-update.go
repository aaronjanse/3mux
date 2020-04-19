package wm

// this is used to update who is in control of the cursor
func (u *Universe) updateSelection() {
	for idx, w := range u.workspaces {
		w.UpdateSelection(idx == u.selectionIdx)
	}
}

func (s *workspace) UpdateSelection(selected bool) {
	s.contents.UpdateSelection(selected)
}

func (s *split) UpdateSelection(selected bool) {
	s.selected = selected
	for idx, n := range s.elements {
		n.contents.UpdateSelection(selected && idx == s.selectionIdx)
	}
}
