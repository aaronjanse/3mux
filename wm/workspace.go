package wm

import (
	"fmt"

	"github.com/aaronjanse/3mux/ecma48"
)

// A workspace is a desktop
type workspace struct {
	contents     *split
	doFullscreen bool
	renderer     ecma48.Renderer

	onDeath    func(error)
	Dead       bool
	newPane    NewPaneFunc
	renderRect Rect
}

func newWorkspace(renderer ecma48.Renderer, u *Universe, onDeath func(error), renderRect Rect, newPane NewPaneFunc) *workspace {
	w := &workspace{
		doFullscreen: false,
		onDeath:      onDeath,
		newPane:      newPane,
		renderer:     renderer,
		renderRect:   renderRect,
	}
	w.contents = newSplit(renderer, u, w.handleChildDeath, renderRect, false, 0, nil, newPane)
	return w
}

func (s *workspace) serialize() string {
	return fmt.Sprintf("Workspace(%s)", s.contents.Serialize())
}

func (s *workspace) setRenderRect(x, y, w, h int) {
	s.renderRect = Rect{X: x, Y: y, W: w, H: h}
	if s.doFullscreen {
		s.getSelectedNode().SetRenderRect(true, x, y, w, h)
	} else {
		s.contents.SetRenderRect(s.doFullscreen, x, y, w, h)
	}
}
