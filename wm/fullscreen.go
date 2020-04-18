package wm

func (u *Universe) ToggleFullscreen() {
	u.workspaces[u.selectionIdx].toggleFullscreen()
}

func (s *workspace) toggleFullscreen() {
	s.setFullscreen(!s.doFullscreen)
}

func (s *workspace) setFullscreen(fullscreen bool) {
	s.doFullscreen = fullscreen
	s.contents.setFullscreen(
		fullscreen,
		s.contents.GetRenderRect().X,
		s.contents.GetRenderRect().Y,
	)
}

func (s *split) setFullscreen(fullscreen bool, w, h int) {
	for idx, n := range s.elements {
		thisOne := fullscreen && idx == s.selectionIdx
		switch child := n.contents.(type) {
		case Container:
			child.setFullscreen(thisOne, w, h)
		case Node:
			if thisOne {
				child.SetRenderRect(fullscreen, 0, 0, w, h)
			} else {
				child.SetPaused(fullscreen)
			}
		}
	}
}

func (s *split) SetPaused(paused bool) {
	for _, n := range s.elements {
		n.contents.SetPaused(paused)
	}
}
