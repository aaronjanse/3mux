package wm

func (u *Universe) SelectAtCoords(x, y int) {
	if u.workspaces[0].doFullscreen {
		return
	}

	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

	u.workspaces[u.selectionIdx].selectAtCoords(x, y)
	u.updateSelection()
	u.drawSelectionBorder()
}

func (s *workspace) selectAtCoords(x, y int) {
	if !s.doFullscreen {
		s.contents.selectAtCoords(x, y)
	}
}

func (s *split) selectAtCoords(x, y int) {
	for idx, n := range s.elements {
		r := n.contents.GetRenderRect()

		vertValid := r.Y <= y && y < r.Y+r.H
		horizValid := r.X <= x && x < r.X+r.W
		if vertValid && horizValid {
			switch child := n.contents.(type) {
			case Container:
				child.selectAtCoords(x, y)
			}
			s.selectionIdx = idx
			return
		}
	}
}

func (u *Universe) DragBorder(x1, y1, x2, y2 int) {
	if u.workspaces[u.selectionIdx].doFullscreen {
		return
	}

	u.workspaces[u.selectionIdx].dragBorder(x1, y1, x2, y2)
	u.redrawAllLines()
	u.drawSelectionBorder()
}

func (s *workspace) dragBorder(x1, y1, x2, y2 int) {
	s.contents.dragBorder(x1, y1, x2, y2)
}

func (s *split) dragBorder(x1, y1, x2, y2 int) {
	for idx, n := range s.elements {
		r := n.contents.GetRenderRect()

		// test if we're at a divider
		horiz := !s.verticallyStacked && x1 == r.X+r.W
		vert := s.verticallyStacked && y1 == r.Y+r.H
		if horiz || vert {
			firstRec := s.elements[idx].contents.GetRenderRect()
			secondRec := s.elements[idx+1].contents.GetRenderRect()

			var combinedSize int
			if s.verticallyStacked {
				combinedSize = firstRec.H + secondRec.H
			} else {
				combinedSize = firstRec.W + secondRec.W
			}

			var wantedRelativeBorderPos int
			if s.verticallyStacked {
				wantedRelativeBorderPos = y2 - firstRec.Y
			} else {
				wantedRelativeBorderPos = x2 - firstRec.X
			}

			wantedBorderRatio := float32(wantedRelativeBorderPos) / float32(combinedSize)
			totalProportion := s.elements[idx].size + s.elements[idx+1].size

			if wantedBorderRatio > 1 { // user did an impossible drag
				return
			}

			s.elements[idx].size = wantedBorderRatio * totalProportion
			s.elements[idx+1].size = (1 - wantedBorderRatio) * totalProportion
			s.refreshRenderRect(false)
			return
		}

		// test if we're within a child
		withinVert := r.Y <= y1 && y1 < r.Y+r.H
		withinHoriz := r.X <= x1 && x1 < r.X+r.W
		if withinVert && withinHoriz {
			switch child := n.contents.(type) {
			case Container:
				child.dragBorder(x1, y1, x2, y2)
			}
			return
		}
	}
}
