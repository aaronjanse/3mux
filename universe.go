package main

// import (
// 	"fmt"
// )

// // A Universe contains workspaces
// type Universe struct {
// 	workspaces   []*Workspace
// 	selectionIdx int

// 	renderRect Rect
// }

// func (u *Universe) simplify() {
// 	for _, e := range u.workspaces {
// 		e.contents.simplify()
// 	}
// }

// func (u *Universe) serialize() string {
// 	out := fmt.Sprintf("Universe[%d]", u.selectionIdx)

// 	out += "("
// 	for i, e := range u.workspaces {
// 		if i != 0 {
// 			out += ", "
// 		}
// 		out += e.serialize()
// 	}
// 	out += ")"

// 	return out
// }

// // setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// // this for when a split is reshaped
// func (u *Universe) setRenderRect(x, y, w, h int) {
// 	u.renderRect = Rect{x, y, w, h}

// 	// NOTE: should we clear the screen?

// 	u.refreshRenderRect()
// }

// func (u *Universe) getRenderRect() Rect {
// 	return u.renderRect
// }

// func (u *Universe) kill() {
// 	for _, n := range u.workspaces {
// 		n.contents.kill()
// 	}
// }

// func (u *Universe) setPause(pause bool) {
// 	for _, n := range u.workspaces {
// 		n.contents.setPause(pause)
// 	}
// }

// // refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// // this is for when one or more of a split's children are reshaped
// func (u *Universe) refreshRenderRect() {
// 	x := u.renderRect.x
// 	y := u.renderRect.y
// 	w := u.renderRect.w
// 	h := u.renderRect.h

// 	for _, child := range u.workspaces {
// 		child.setRenderRect(x, y, w, h)
// 	}
// }
