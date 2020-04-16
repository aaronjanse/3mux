package main

import (
	"fmt"
)

// A Workspace is a desktop
type Workspace struct {
	contents     *Split
	doFullscreen bool
}

func (s *Workspace) serialize() string {
	return fmt.Sprintf("Workspace(%s)", s.contents.serialize())
}

func (s *Workspace) setRenderRect(x, y, w, h int) {
	if s.doFullscreen {
		getSelection().getContainer().setRenderRect(x, y, w, h)
	} else {
		s.contents.setRenderRect(x, y, w, h)
	}
}

func (s *Workspace) addPane() {
	s.contents.addPane()
}

func (s *Workspace) selectAtCoords(x, y int) {
	s.contents.selectAtCoords(x, y)
}

func (s *Workspace) updateSelection(selected bool) {
	s.contents.updateSelection(selected)
}

func (s *Workspace) dragBorder(x1, y1, x2, y2 int) {
	s.contents.dragBorder(x1, y1, x2, y2)
}
