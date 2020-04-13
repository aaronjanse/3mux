package main

// import (
// 	"fmt"
// )

// // A Workspace is a desktop
// type Workspace struct {
// 	contents     *Split
// 	doFullscreen bool
// }

// func (s *Workspace) serialize() string {
// 	return fmt.Sprintf("Workspace(%s)", s.contents.serialize())
// }

// func (s *Workspace) setRenderRect(x, y, w, h int) {
// 	if s.doFullscreen {
// 		getSelection().getContainer().setRenderRect(x, y, w, h)
// 	} else {
// 		s.contents.setRenderRect(x, y, w, h)
// 	}
// }
