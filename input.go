package main

import (
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/wm"
)

type inputState struct {
	mouseDown bool
}

const demoMode = false

var demoTextTimer *time.Timer = nil
var demoTextDuration = 1000 * time.Millisecond

// handleInput puts the input through a series of switches and seive functions.
// When something acts on the event, we stop passing it downstream
func handleInput(u *wm.Universe, human string, obj ecma48.Output) {
	if demoMode {
		renderer.DemoText = human

		if demoTextTimer == nil {
			demoTextTimer = time.NewTimer(demoTextDuration)
		} else {
			demoTextTimer.Stop()
			demoTextTimer.Reset(demoTextDuration)
		}

		go func() {
			<-demoTextTimer.C
			renderer.DemoText = ""
		}()
	}

	if seiveResizeEvents(u, human, obj) {
		return
	}

	if seiveMouseEvents(u, human, obj) {
		return
	}

	if seiveConfigEvents(u, human) {
		return
	}

	// log.Printf("%q %+v", obj.Raw, obj.Parsed)
	// t := getSelection().getContainer().(*Pane)

	switch x := obj.Parsed.(type) {
	case ecma48.CursorMovement:
		switch x.Direction {
		case ecma48.Up:
			u.HandleStdin("\x1bOA")
		case ecma48.Down:
			u.HandleStdin("\x1bOB")
		case ecma48.Right:
			u.HandleStdin("\x1bOC")
		case ecma48.Left:
			u.HandleStdin("\x1bOD")
		}
	default:
		// if we didn't find anything special, just pass the raw data to the selected terminal
		u.HandleStdin(string(obj.Raw))
	}

}

func seiveResizeEvents(u *wm.Universe, human string, obj ecma48.Output) bool {
	if resizeMode {
		switch human {
		case "Up", "k":
			u.ResizePane(wm.Up)
		case "Down", "j":
			u.ResizePane(wm.Down)
		case "Left", "h":
			u.ResizePane(wm.Left)
		case "Right", "l":
			u.ResizePane(wm.Right)
		default:
			resizeMode = false
		}
		return true
	}
	return false
}

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
