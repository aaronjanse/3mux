package wm

import "github.com/aaronjanse/3mux/render"

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
	HandleStdin(string)
}

type Container interface {
	addPane()
	killPane()
	setFullscreen(fullscreen bool, x, y int)
	selectAtCoords(x, y int)
	dragBorder(x1, y1, x2, y2 int)
	moveWindow(d Direction) (bubble bool, p Node)
	simplify()
	resizePane(d Direction) (bubble bool)
	moveSelection(d Direction) (bubble bool)
	cycleSelection(forwards bool) (bubble bool)
	selectMin()
	selectMax()
	Node
}

type NewPaneFunc func(*render.Renderer) Node
