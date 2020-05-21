package main

import (
	"fmt"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/wm"
)

var mouseDownX, mouseDownY int

// seiveMouseEvents processes mouse events and returns true if the data should *not* be passed downstream
func seiveMouseEvents(u *wm.Universe, human string, obj ecma48.Output) bool {
	switch ev := obj.Parsed.(type) {
	case ecma48.MouseDown:
		u.SelectAtCoords(ev.X, ev.Y)
		mouseDownX = ev.X
		mouseDownY = ev.Y
	case ecma48.MouseUp:
		u.DragBorder(mouseDownX, mouseDownY, ev.X, ev.Y)
	case ecma48.MouseDrag:
		// do nothing
	case ecma48.ScrollUp:
		u.ScrollUp()
	case ecma48.ScrollDown:
		u.ScrollDown()
	default:
		return false
	}

	return true
}

func humanify(obj ecma48.Output) string {
	humanCode := ""
	switch x := obj.Parsed.(type) {
	case ecma48.Char:
		humanCode = string(x.Rune)
	case ecma48.CtrlChar:
		humanCode = fmt.Sprintf("Ctrl+%s", humanifyRune(x.Char))
	case ecma48.AltChar:
		humanCode = fmt.Sprintf("Alt+%s", humanifyRune(x.Char))
	case ecma48.AltShiftChar:
		humanCode = fmt.Sprintf("Alt+Shift+%s", humanifyRune(x.Char))
	case ecma48.Tab:
		humanCode = "Tab"
	case ecma48.CursorMovement:
		if x.Ctrl {
			humanCode += "Ctrl+"
		}
		if x.Alt {
			humanCode += "Alt+"
		}
		if x.Shift {
			humanCode += "Shift+"
		}
		switch x.Direction {
		case ecma48.Up:
			humanCode += "Up"
		case ecma48.Down:
			humanCode += "Down"
		case ecma48.Left:
			humanCode += "Left"
		case ecma48.Right:
			humanCode += "Right"
		}
	}
	return humanCode
}

func humanifyRune(r rune) string {
	switch r {
	case '\n', '\r':
		return "Enter"
	default:
		return string(r)
	}
}
