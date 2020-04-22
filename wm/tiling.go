package wm

import (
	"github.com/aaronjanse/3mux/ecma48"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	X, Y, W, H int
}

// A SizedNode represents a single pane of a split, having a size (relative to a total 1.0) and renderable contents
type SizedNode struct {
	size     float32
	contents Node
}

// Direction is the type of Up, Down, Left, and Right
type Direction int

// directions
const (
	_ Direction = iota
	Up
	Down
	Left
	Right
)

type Node interface {
	SetRenderRect(fullscreen bool, x, y, w, h int)
	GetRenderRect() Rect
	Serialize() string
	SetPaused(bool)
	SetDeathHandler(func(error))
	Kill()
	IsDead() bool
	UpdateSelection(selected bool)
	ToggleSearch()
	ScrollUp()
	ScrollDown()
	HandleStdin(ecma48.Output)
}

type Container interface {
	addPane()
	killPane() bool
	setFullscreen(fullscreen bool, x, y int)
	selectAtCoords(x, y int)
	dragBorder(x1, y1, x2, y2 int)
	moveWindow(d Direction) (bubble bool, superBubble bool, p Node)
	simplify()
	resizePane(d Direction) (bubble bool)
	moveSelection(d Direction) (bubble bool)
	cycleSelection(forwards bool) (bubble bool)
	selectMin()
	selectMax()
	getSelectedNode() Node
	addPaneTmux(vert bool)
	Node
}

type NewPaneFunc func(ecma48.Renderer) Node

var FuncNames = map[string]func(*Universe){
	"new-pane":  func(u *Universe) { u.AddPane() },
	"kill-pane": func(u *Universe) { u.KillPane() },

	"split-pane-horiz": func(u *Universe) { u.AddPaneTmux(false) },
	"split-pane-vert":  func(u *Universe) { u.AddPaneTmux(true) },

	"show-help":     func(u *Universe) {},
	"hide-help-bar": func(u *Universe) {},

	"toggle-fullscreen": func(u *Universe) { u.ToggleFullscreen() },
	"toggle-search":     func(u *Universe) { u.ToggleSearch() },

	"resize-up":    func(u *Universe) { u.ResizePane(Up) },
	"resize-down":  func(u *Universe) { u.ResizePane(Down) },
	"resize-left":  func(u *Universe) { u.ResizePane(Left) },
	"resize-right": func(u *Universe) { u.ResizePane(Right) },

	"move-pane-up":    func(u *Universe) { u.MoveWindow(Up) },
	"move-pane-down":  func(u *Universe) { u.MoveWindow(Down) },
	"move-pane-left":  func(u *Universe) { u.MoveWindow(Left) },
	"move-pane-right": func(u *Universe) { u.MoveWindow(Right) },

	"move-selection-up":    func(u *Universe) { u.MoveSelection(Up) },
	"move-selection-down":  func(u *Universe) { u.MoveSelection(Down) },
	"move-selection-left":  func(u *Universe) { u.MoveSelection(Left) },
	"move-selection-right": func(u *Universe) { u.MoveSelection(Right) },
}
