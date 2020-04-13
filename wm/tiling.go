package wm

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

// A Node represents a single pane of a split, having a size (relative to a total 1.0) and renderable contents
type Node struct {
	size     float32
	contents Container
}

// A Container is a renderable, debuggable, and killable unit of the window tree
type Container interface {
	// setRenderRect(x, y, w, h int)
	Reshape(x, y, w, h int)
	Kill()
	// getRenderRect() Rect
	Serialize() string
	// simplify()
	// kill()
	// setPause(bool)
}

var root Universe
