package main

import (
	"github.com/aaronjanse/i3-tmux/keypress"
)

type inputState struct {
	mouseDown bool
}

// handleInput puts the input through a series of switches and seive functions.
// When something acts on the event, we stop passing it downstream
func handleInput(event interface{}, rawData []byte) {
	defer func() {
		if config.statusBar {
			debug(root.serialize())
		}
	}()

	if resizeMode {
		switch ev := event.(type) {
		case keypress.Arrow:
			// we can get rid of this if keypress and everything else have a common Direction enum
			switch ev.Direction {
			case keypress.Up:
				resizeWindow(Up, 0.1)
			case keypress.Down:
				resizeWindow(Down, 0.1)
			case keypress.Left:
				resizeWindow(Left, 0.1)
			case keypress.Right:
				resizeWindow(Right, 0.1)
			}
		case keypress.Character:
			switch ev.Char {
			case 'i':
				resizeWindow(Up, 0.1)
			case 'k':
				resizeWindow(Down, 0.1)
			case 'j':
				resizeWindow(Left, 0.1)
			case 'l':
				resizeWindow(Right, 0.1)
			}
		case keypress.Enter:
			resizeMode = false
		}
		return
	}

	switch ev := event.(type) {
	case keypress.Resize:
		resize(ev.W, ev.H)
		return
	}

	if seiveMouseEvents(event) {
		return
	}

	if seiveConfigEvents(event) {
		return
	}

	// if we didn't find anything special, just pass the raw data to the selected terminal

	t := getSelection().getContainer().(*Pane)

	t.shell.handleStdin(string(rawData))
	t.vterm.RefreshCursor()
}

// seiveMouseEvents processes mouse events and returns true if the data should *not* be passed downstream
func seiveMouseEvents(event interface{}) bool {
	switch ev := event.(type) {
	case keypress.MouseDown:
		// are we clicking a border? if so, which one?
		path := findClosestBorderForCoord([]int{}, ev.X, ev.Y)
		pane := path.getContainer()
		r := pane.getRenderRect()

		if ev.X == r.x+r.w+1 || ev.Y == r.y+r.h+1 {
			mouseDownPath = path
		}
	case keypress.MouseUp:
		if mouseDownPath != nil { // end resize
			x := ev.X
			y := ev.Y

			parent, _ := mouseDownPath.getParent()
			pane := mouseDownPath.getContainer()
			pr := pane.getRenderRect()
			sr := parent.getRenderRect()

			var desiredArea int
			var proportionOfParent float32
			if parent.verticallyStacked {
				desiredArea = y - pr.y - 1
				proportionOfParent = float32(desiredArea) / float32(sr.h)
			} else {
				desiredArea = x - pr.x - 1
				proportionOfParent = float32(desiredArea) / float32(sr.w)
			}

			focusIdx := mouseDownPath[len(mouseDownPath)-1]
			currentProportion := parent.elements[focusIdx].size
			numElementsFollowing := len(parent.elements) - (focusIdx + 1)
			avgShift := (proportionOfParent - currentProportion) / float32(numElementsFollowing)
			for i := focusIdx + 1; i < len(parent.elements); i++ {
				parent.elements[i].size += avgShift
			}

			parent.elements[focusIdx].size = proportionOfParent

			parent.refreshRenderRect()

			mouseDownPath = nil
		}
	case keypress.ScrollUp:
		t := getSelection().getContainer().(*Pane)
		t.vterm.ScrollbackDown()
	case keypress.ScrollDown:
		t := getSelection().getContainer().(*Pane)
		t.vterm.ScrollbackUp()
	default:
		return false
	}

	return true
}
