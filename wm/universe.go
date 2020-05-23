package wm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aaronjanse/3mux/ecma48"
)

// A Universe contains workspaces
type Universe struct {
	workspaces   []*workspace
	selectionIdx int
	renderRect   Rect
	renderer     ecma48.Renderer

	onDeath func(error)
	dead    bool

	helpBar         bool
	enableStatusBar bool

	wmOpMutex *sync.Mutex
}

func NewUniverse(renderer ecma48.Renderer, helpBar bool, enableStatusBar bool, onDeath func(error), renderRect Rect, newPane NewPaneFunc) *Universe {
	u := &Universe{
		selectionIdx:    0,
		renderRect:      renderRect,
		onDeath:         onDeath,
		renderer:        renderer,
		helpBar:         helpBar,
		enableStatusBar: enableStatusBar,
		wmOpMutex:       &sync.Mutex{},
	}
	u.workspaces = []*workspace{newWorkspace(renderer, u, u.handleChildDeath, renderRect, newPane)}
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

func (u *Universe) IsDead() bool {
	return u.dead
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

func (u *Universe) redrawAllLines() {
	for _, n := range u.workspaces {
		n.redrawAllLines()
	}
}

func (s *workspace) redrawAllLines() {
	if !s.doFullscreen {
		s.contents.redrawLines()
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
		if u.helpBar {
			child.setRenderRect(x, y, w, h-2)
		} else if u.enableStatusBar {
			child.setRenderRect(x, y, w, h-1)
		} else {
			child.setRenderRect(x, y, w, h)
		}
	}

	if u.helpBar {
		u.drawHelpBar()
	} else if u.enableStatusBar {
		u.drawStatusBar()
	}

	u.redrawAllLines()
	u.drawSelectionBorder()
}

func (u *Universe) drawStatusBar() {
	text := "3mux"
	for i := 0; i < u.renderRect.W; i++ {
		var r rune
		if i < len(text) {
			r = rune(text[i])
		} else {
			r = 0
		}

		ch := ecma48.PositionedChar{
			Rune: r,
			Cursor: ecma48.Cursor{
				X: i, Y: u.renderRect.H - 1,
				Style: ecma48.Style{
					Fg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Normal,
						Code:      0,
					},
					Bg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Normal,
						Code:      2,
					},
				},
			},
		}

		u.renderer.HandleCh(ch)
	}
}

func (u *Universe) drawHelpBar() {
	for _, hb := range helpBar {
		if helpBarMinLen(hb[0]) > u.renderRect.W {
			continue
		}

		widthStr := hb[0]
		widthStr = strings.ReplaceAll(widthStr, "{", "")
		widthStr = strings.ReplaceAll(widthStr, "}", "")

		space := " "
		for {
			test := strings.ReplaceAll(widthStr, "\t", space+" ")
			if len(test) >= u.renderRect.W {
				break
			}
			space += " "
		}

		style := ecma48.Style{}

		for line := 0; line < 2; line++ {
			x := 0
			for _, r := range strings.ReplaceAll(hb[line], "\t", space) {
				// log.Printf("%q", r)
				switch r {
				case '{':
					style.Reverse = true
				case '}':
					style.Reverse = false
				default:
					u.renderer.HandleCh(ecma48.PositionedChar{
						Rune: r,
						Cursor: ecma48.Cursor{
							X: x, Y: u.renderRect.H - 2 + line,
							Style: style,
						},
					})
					x++
				}
			}
		}

		break
	}
}

func (u *Universe) HideHelpBar() {
	u.helpBar = false
	u.refreshRenderRect()
}

func helpBarMinLen(str string) int {
	var x string
	x = strings.ReplaceAll(str, "{", "")
	x = strings.ReplaceAll(str, "}", "")
	x = strings.ReplaceAll(str, "\t", " ")
	return len(x)
}

var helpBar [][2]string = [][2]string{
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Move Selection\t{/} Search    \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Move Pane     \t{F} Fullscreen",
	},
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Move Selection\t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Move Pane",
	},
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Select\t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Move",
	},
	[2]string{
		"Alt+...      \t{N} New  \t{Arrow} Select\t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close\t{Arrow} Move",
	},
	[2]string{
		"Alt+...      \t{N} New  \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close",
	},
	[2]string{
		"",
		"{Alt+X}       Hide Help",
	},
	[2]string{
		"",
		"",
	},
}
