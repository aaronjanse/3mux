package wm

import (
	"fmt"
	"log"

	"github.com/aaronjanse/3mux/pane"
	gc "github.com/rthornton128/goncurses"
)

// A Universe contains workspaces
type Universe struct {
	workspaces   []*Workspace
	selectionIdx int

	renderRect Rect
}

func NewUniverse(stdscr *gc.Window, w, h int) *Universe {
	pane, err := pane.NewPane(0, 0, w, h, func() {})
	if err != nil {
		log.Fatal(err)
	}
	return &Universe{
		workspaces: []*Workspace{
			&Workspace{
				contents: &Split{
					renderRect: Rect{
						x: 0,
						y: 0,
						w: w,
						h: h,
					},
					stdscr:            stdscr,
					verticallyStacked: false,
					selectionIdx:      0,
					elements: []Node{
						Node{
							size:     1,
							contents: pane, // FIXME onDeath missing
						},
					}},
				doFullscreen: false,
			},
		},
		selectionIdx: 0,
	}
}

// func (u *Universe) Simplify() {
// 	for _, e := range u.workspaces {
// 		e.contents.Simplify()
// 	}
// }

func (u *Universe) Serialize() string {
	out := fmt.Sprintf("Universe[%d]", u.selectionIdx)

	out += "("
	for i, e := range u.workspaces {
		if i != 0 {
			out += ", "
		}
		out += e.Serialize()
	}
	out += ")"

	return out
}

// Reshape updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (u *Universe) Reshape(x, y, w, h int) {
	u.renderRect = Rect{x, y, w, h}

	for _, child := range u.workspaces {
		child.Reshape(x, y, w, h)
	}
}

func (u *Universe) getRenderRect() Rect {
	return u.renderRect
}

func (u *Universe) Kill() {
	for _, n := range u.workspaces {
		n.contents.Kill()
	}
}

func (u *Universe) AddPane() {
	log.Println("ADDING")
	u.workspaces[u.selectionIdx].AddPane()
	log.Println(u.Serialize())
}

func (u *Universe) HandleStdin(in string) {
	// log.Println("IN", []byte(in))
	// log.Println("sel", u.getSelection())
	u.getSelection().getContainer(u).(*pane.Pane).HandleStdin(in)
}

// func (u *Universe) setPause(pause bool) {
// 	for _, n := range u.workspaces {
// 		n.contents.setPause(pause)
// 	}
// }

func (u *Universe) UpdateFocus() {
	u.getSelection().getContainer(u).(*pane.Pane).RefreshCursor()
}
