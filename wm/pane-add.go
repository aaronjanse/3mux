package wm

import (
	"errors"
)

func (u *Universe) AddPane() error {
	err := u.workspaces[u.selectionIdx].addPane()
	if err != nil {
		return err
	}
	u.refreshRenderRect() // FIXME only needs to redraw lines!
	u.updateSelection()
	return nil
}

func (s *workspace) addPane() error {
	if s.doFullscreen {
		return errors.New("cannot add pane while one is fullscreen")
	}
	s.contents.addPane()
	return nil
}

func (s *split) addPane() {
	switch x := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		x.addPane()
	case Node:
		if len(s.elements) > 8 {
			return
		}

		size := float32(1) / float32(len(s.elements)+1)

		// resize siblings
		scaleFactor := float32(1) - size
		for i := range s.elements {
			s.elements[i].size *= scaleFactor
		}

		// add new child
		createdTerm := s.newPane(s.renderer)
		createdTerm.SetDeathHandler(s.handleChildDeath)
		s.elements = append(s.elements, SizedNode{
			size:     size,
			contents: createdTerm,
		})

		// update selection to new child
		s.selectionIdx = len(s.elements) - 1
	}
}

func (u *Universe) AddPaneTmux(vert bool) error {
	err := u.workspaces[u.selectionIdx].addPane()
	if err != nil {
		return err
	}
	u.refreshRenderRect() // FIXME only needs to redraw lines!
	u.updateSelection()
	return nil
}

func (s *workspace) addPaneTmux(vert bool) error {
	if s.doFullscreen {
		return errors.New("cannot add pane while one is fullscreen")
	}
	s.contents.addPane()
	return nil
}

func (s *split) addPaneTmux(vert bool) {
	switch x := s.elements[s.selectionIdx].contents.(type) {
	case Container:
		x.addPane()
	case Node:
		s.elements[s.selectionIdx].contents = newSplit(
			s.renderer, s.redrawLines, s.handleChildDeath, x.GetRenderRect(), vert,
			1, []Node{x, s.newPane(s.renderer)}, s.newPane,
		)
		s.refreshRenderRect(false)
	}
}
