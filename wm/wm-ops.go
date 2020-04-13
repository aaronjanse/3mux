package wm

// import (
// 	"github.com/aaronjanse/3mux/keypress"
// )

// func search() {
// 	getSelection().getContainer().(*Pane).toggleSearch()
// }

// func fullscreen() {
// 	root.workspaces[root.selectionIdx].doFullscreen = true
// 	path := getSelection()

// 	root.workspaces[root.selectionIdx].contents.setPause(true)

// 	r := root.workspaces[root.selectionIdx].contents.getRenderRect()
// 	specaialPane := path.getContainer()
// 	specaialPane.setRenderRect(r.x, r.y, r.w, r.h)
// 	specaialPane.setPause(false)

// 	keypress.ShouldProcessMouse(false)
// }

// func unfullscreen() {
// 	root.workspaces[root.selectionIdx].doFullscreen = false
// 	root.workspaces[root.selectionIdx].contents.setPause(false)

// 	keypress.ShouldProcessMouse(true)

// 	root.refreshRenderRect()
// }

func (u *Universe) MoveWindow(d Direction) {
	path := u.getSelection()
	parent, parentPath := path.getParent(u)

	root := u

	vert := parent.verticallyStacked

	rootRect := root.workspaces[root.selectionIdx].contents.renderRect

	if (!vert && d == Left) || (vert && d == Up) {
		idx := parent.selectionIdx

		if idx == 0 {
			if len(parentPath) < 2 {
				return
			}

			grandparent, grandparentPath := parentPath.getParent(u)
			tmp := parentPath.popContainer(u, parent.selectionIdx)

			if len(parentPath) == 2 {
				root.workspaces[root.selectionIdx].contents = &Split{
					lock:              u.lock,
					renderRect:        Rect{w: rootRect.w, h: rootRect.h},
					verticallyStacked: vert,
					stdscr:            u.stdscr,
					selectionIdx:      1,
					elements: []Node{
						Node{
							size:     0.5,
							contents: tmp,
						},
						Node{
							size:     0.5,
							contents: grandparent,
						},
					},
				}
			} else {
				greatGrandparent, _ := grandparentPath.getParent(u)
				greatGrandparent.insertContainer(tmp, greatGrandparent.selectionIdx)
			}

			// root.workspaces[root.selectionIdx].contents.refreshChildShapes()
		} else {
			tmp := parent.elements[idx-1]
			parent.elements[idx-1] = parent.elements[idx]
			parent.elements[idx] = tmp

			parent.selectionIdx--
		}

		// root.refreshRenderRect()
	} else if (!vert && d == Right) || (vert && d == Down) {
		idx := parent.selectionIdx

		if idx == len(parent.elements)-1 {
			if len(parentPath) < 2 {
				return
			}

			grandparent, grandparentPath := parentPath.getParent(u)
			tmp := parentPath.popContainer(u, parent.selectionIdx)

			if len(parentPath) == 2 {
				root.workspaces[root.selectionIdx].contents = &Split{
					lock:              u.lock,
					renderRect:        Rect{w: rootRect.w, h: rootRect.h},
					verticallyStacked: vert,
					selectionIdx:      1,
					stdscr:            u.stdscr,
					elements: []Node{
						Node{
							size:     0.5,
							contents: grandparent,
						},
						Node{
							size:     0.5,
							contents: tmp,
						},
					},
				}
			} else {
				greatGrandparent, _ := grandparentPath.getParent(u)
				greatGrandparent.insertContainer(tmp, grandparent.selectionIdx+2)
				greatGrandparent.selectionIdx++
			}

			// root.workspaces[root.selectionIdx].contents.refreshChildShapes()
		} else {
			tmp := parent.elements[idx+1]
			parent.elements[idx+1] = parent.elements[idx]
			parent.elements[idx] = tmp

			parent.selectionIdx++
		}

		// root.refreshRenderRect()
	} else {
		movingVert := d == Up || d == Down

		p := path
		for len(p) > 1 {
			s, _ := p.getParent(u)
			if s.verticallyStacked == movingVert {
				tmp := parentPath.popContainer(u, parent.selectionIdx)

				if d == Left || d == Up {
					s.insertContainer(tmp, s.selectionIdx)
				} else {
					s.insertContainer(tmp, s.selectionIdx+1)
					s.selectionIdx++
				}

				// root.refreshRenderRect()
				break
			}
			p = p[:len(p)-1]
		}

		if len(p) == 1 && len(parent.elements) > 1 {
			tmp := parentPath.popContainer(u, parent.selectionIdx)
			tmpRoot := root.workspaces[root.selectionIdx].contents

			root.workspaces[root.selectionIdx].contents = &Split{
				lock:              u.lock,
				renderRect:        Rect{w: rootRect.w, h: rootRect.h},
				verticallyStacked: movingVert,
				stdscr:            u.stdscr,
				selectionIdx:      0,
				elements: []Node{
					Node{
						size:     1,
						contents: tmpRoot,
					},
				},
			}

			insertIdx := 0
			if d == Down || d == Right {
				insertIdx = 1
			}
			root.workspaces[root.selectionIdx].contents.insertContainer(tmp, insertIdx)
			root.workspaces[root.selectionIdx].contents.selectionIdx = insertIdx

			// root.refreshRenderRect()
		}
	}

	// parent.refreshChildShapes()
	u.workspaces[root.selectionIdx].contents.refreshChildShapes()
	u.UpdateFocus()
}

// func killWindow() {
// 	parent, parentPath := getSelection().getParent()
// 	t := parentPath.popContainer(parent.selectionIdx)
// 	t.(*Pane).kill()

// 	// FIXME: allows for only one workspace
// 	if len(root.workspaces[root.selectionIdx].contents.elements) == 0 {
// 		shutdownNow()
// 		return
// 	}

// 	// select the new Term
// 	newTerm := getSelection().getContainer().(*Pane)
// 	newTerm.selected = true
// 	newTerm.softRefresh()
// 	newTerm.vterm.RefreshCursor()

// 	if len(root.workspaces[root.selectionIdx].contents.elements) == 1 {
// 		keypress.ShouldProcessMouse(false)
// 	}
// }

// // stuff like h(h(x), y) -> h(x, y)
// func (s *Split) simplify() {
// 	if len(s.elements) == 1 {
// 		switch child := (*s).elements[0].contents.(type) {
// 		case *Split:
// 			s.verticallyStacked = child.verticallyStacked
// 			s.elements = child.elements
// 			s.selectionIdx = child.selectionIdx
// 		}
// 	} else {
// 		newElements := []Node{}
// 		selectionIdx := (*s).selectionIdx
// 		for idx, n := range (*s).elements {
// 			switch child := n.contents.(type) {
// 			case *Split:
// 				if child.verticallyStacked == s.verticallyStacked {
// 					for j := range child.elements {
// 						child.elements[j].size *= n.size
// 					}
// 					newElements = append(newElements, child.elements...)
// 					if idx == s.selectionIdx {
// 						selectionIdx += child.selectionIdx
// 					} else if idx < s.selectionIdx {
// 						selectionIdx += len(child.elements) - 1
// 					}
// 				} else {
// 					newElements = append(newElements, n)
// 				}
// 			case *Pane:
// 				newElements = append(newElements, n)
// 			}
// 		}
// 		s.elements = newElements
// 		s.selectionIdx = selectionIdx
// 	}

// 	for _, n := range s.elements {
// 		switch child := n.contents.(type) {
// 		case *Split:
// 			child.simplify()
// 		}
// 	}
// }

func (u *Universe) MoveSelection(d Direction) {
	path := u.getSelection()

	parent, _ := path.getParent(u)

	vert := parent.verticallyStacked

	if (d == Left && !vert) || (d == Up && vert) {
		parent.selectionIdx--
		if parent.selectionIdx < 0 {
			parent.selectionIdx = 0
		}
	} else if (d == Right && !vert) || (d == Down && vert) {
		parent.selectionIdx++
		if parent.selectionIdx > len(parent.elements)-1 {
			parent.selectionIdx = len(parent.elements) - 1
		}
	} else {
		movingVert := d == Up || d == Down

		p := path
		for len(p) > 1 {
			s, _ := p.getParent(u)
			if s.verticallyStacked == movingVert {
				if d == Up || d == Left {
					s.selectionIdx--
					if s.selectionIdx < 0 {
						s.selectionIdx = 0
					}
				} else {
					s.selectionIdx++
					if s.selectionIdx > len(s.elements)-1 {
						s.selectionIdx = len(s.elements) - 1
					}
				}
				break
			}
			p = p[:len(p)-1]
		}
	}

	u.UpdateFocus()
}

// func resizeWindow(d Direction, diff float32) {
// 	resizeWindowImpl(getSelection(), d, diff)
// }

// func resizeWindowImpl(path Path, d Direction, diff float32) {
// 	parent, parentPath := path.getParent()

// 	const shift = 0.1
// 	var delta float32

// 	if parent.verticallyStacked {
// 		if d == Left || d == Right {
// 			resizeWindowImpl(parentPath, d, diff)
// 			return
// 		}

// 		if d == Up {
// 			delta = +shift
// 		} else if d == Down {
// 			delta = -shift
// 		}
// 	} else {
// 		if d == Up || d == Down {
// 			resizeWindowImpl(parentPath, d, diff)
// 			return
// 		}

// 		if d == Right {
// 			delta = +shift
// 		} else if d == Left {
// 			delta = -shift
// 		}
// 	}

// 	size := parent.elements[parent.selectionIdx].size

// 	selContents := parent.elements[parent.selectionIdx].contents

// 	var wh float32
// 	if parent.verticallyStacked {
// 		wh = float32(selContents.getRenderRect().h)
// 	} else {
// 		wh = float32(selContents.getRenderRect().w)
// 	}

// 	var min float32 = 1
// 	if (size+delta)*wh <= min {
// 		return
// 	}

// 	scale := (float32(1) - size - delta) / (float32(1) - size)

// 	for i := range parent.elements {
// 		if i != parent.selectionIdx {
// 			elem := parent.elements[i]
// 			newSize := elem.size * scale

// 			var wh float32
// 			if parent.verticallyStacked {
// 				wh = float32(elem.contents.getRenderRect().h)
// 			} else {
// 				wh = float32(elem.contents.getRenderRect().w)
// 			}

// 			if newSize*wh <= min {
// 				return
// 			}
// 		}
// 	}

// 	parent.elements[parent.selectionIdx].size += delta
// 	for i := range parent.elements {
// 		if i != parent.selectionIdx {
// 			parent.elements[i].size *= scale
// 		}
// 	}

// 	root.refreshRenderRect()
// }
