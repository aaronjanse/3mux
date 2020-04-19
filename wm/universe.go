package wm

import (
	"fmt"

	"github.com/aaronjanse/3mux/render"
)

// A Universe contains workspaces
type Universe struct {
	workspaces   []*workspace
	selectionIdx int
	renderRect   Rect

	onDeath func(error)
}

func NewUniverse(renderer *render.Renderer, onDeath func(error), renderRect Rect, newPane NewPaneFunc) *Universe {
	u := &Universe{
		selectionIdx: 0,
		renderRect:   renderRect,
		onDeath:      onDeath,
	}
	u.workspaces = []*workspace{newWorkspace(renderer, u.handleChildDeath, renderRect, newPane)}
	u.updateSelection()
	u.refreshRenderRect()
	return u
}

func (u *Universe) Serialize() string {
	out := fmt.Sprintf("Universe[%d]", u.selectionIdx)

	out += "("
	for i, e := range u.workspaces {
		if i != 0 {
			out += ", "
		}
		out += e.serialize()
	}
	out += ")"

	return out
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (u *Universe) SetRenderRect(x, y, w, h int) {
	u.renderRect = Rect{x, y, w, h}

	// NOTE: should we clear the screen?

	u.refreshRenderRect()
}

func (u *Universe) getRenderRect() Rect {
	return u.renderRect
}

func (u *Universe) setPaused(pause bool) {
	for _, n := range u.workspaces {
		n.contents.SetPaused(pause)
	}
}

// refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// this is for when one or more of a split's children are reshaped
func (u *Universe) refreshRenderRect() {
	x := u.renderRect.X
	y := u.renderRect.Y
	w := u.renderRect.W
	h := u.renderRect.H

	for _, child := range u.workspaces {
		child.setRenderRect(x, y, w, h)
	}
}
