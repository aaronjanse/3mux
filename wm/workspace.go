package wm

import (
	"fmt"

	"github.com/aaronjanse/3mux/render"
)

// A workspace is a desktop
type workspace struct {
	contents     *split
	doFullscreen bool

	onDeath func(error)
	Dead    bool
	newPane NewPaneFunc
}

func newWorkspace(renderer *render.Renderer, onDeath func(error), renderRect Rect, newPane NewPaneFunc) *workspace {
	w := &workspace{
		doFullscreen: false,
		onDeath:      onDeath,
		newPane:      newPane,
	}
	w.contents = newSplit(renderer, w.handleChildDeath, renderRect, false, nil, newPane)
	return w
}

func (s *workspace) serialize() string {
	return fmt.Sprintf("Workspace(%s)", s.contents.Serialize())
}

func (s *workspace) setRenderRect(x, y, w, h int) {
	s.contents.SetRenderRect(s.doFullscreen, x, y, w, h)
}
