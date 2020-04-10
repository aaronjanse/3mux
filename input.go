package main

import (
	"time"

	"github.com/aaronjanse/3mux/keypress"
)

type inputState struct {
	mouseDown bool
}

const demoMode = false

var demoTextTimer *time.Timer = nil
var demoTextDuration = 1000 * time.Millisecond

// handleInput puts the input through a series of switches and seive functions.
// When something acts on the event, we stop passing it downstream
func handleInput(event interface{}, rawData []byte) {
	if demoMode {
		switch ev := event.(type) {
		case keypress.AltChar:
			renderer.DemoText = "Alt + " + string(ev.Char)
		case keypress.AltShiftChar:
			renderer.DemoText = "Alt + Shift + " + string(ev.Char)
		case keypress.AltArrow:
			switch ev.Direction {
			case keypress.Up:
				renderer.DemoText = "Alt + " + string("↑")
			case keypress.Down:
				renderer.DemoText = "Alt + " + string("↓")
			case keypress.Left:
				renderer.DemoText = "Alt + " + string("←")
			case keypress.Right:
				renderer.DemoText = "Alt + " + string("→")
			}
		case keypress.AltShiftArrow:
			switch ev.Direction {
			case keypress.Up:
				renderer.DemoText = "Alt + Shift + " + string("↑")
			case keypress.Down:
				renderer.DemoText = "Alt + Shift + " + string("↓")
			case keypress.Left:
				renderer.DemoText = "Alt + Shift + " + string("←")
			case keypress.Right:
				renderer.DemoText = "Alt + Shift + " + string("→")
			}
		}

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

	defer func() {
		if config.statusBar {
			debug(root.serialize())
		}
	}()

	if resizeMode {
		switch ev := event.(type) {
		case keypress.Direction:
			// we can get rid of this if keypress and everything else have a common Direction enum
			switch ev {
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
			case 'h':
				resizeWindow(Left, 0.1)
			case 'j':
				resizeWindow(Down, 0.1)
			case 'k':
				resizeWindow(Up, 0.1)
			case 'l':
				resizeWindow(Right, 0.1)
			default:
				resizeMode = false
			}
		default:
			resizeMode = false
		}
		if resizeMode {
			return
		}
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

	t.handleStdin(string(rawData))
}

var mouseDownPath Path

// seiveMouseEvents processes mouse events and returns true if the data should *not* be passed downstream
func seiveMouseEvents(event interface{}) bool {
	switch ev := event.(type) {
	case keypress.MouseDown:
		// are we clicking a border? if so, which one?
		path := findClosestBorderForCoord([]int{root.selectionIdx}, ev.X, ev.Y)
		pane := path.getContainer()
		r := pane.getRenderRect()

		if ev.Y == r.y+r.h+1 {
			mouseDownPath = path
			parent, _ := mouseDownPath.getParent()
			if !parent.verticallyStacked {
				mouseDownPath = mouseDownPath[:len(mouseDownPath)-1]
			}
		} else if ev.X == r.x+r.w+1 {
			mouseDownPath = path
			parent, _ := mouseDownPath.getParent()
			if parent.verticallyStacked {
				mouseDownPath = mouseDownPath[:len(mouseDownPath)-1]
			}
		} else {
			// deselect the old Term
			oldTerm := getSelection().getContainer().(*Pane)
			oldTerm.selected = false
			// oldTerm.softRefresh()

			setSelection(path)

			// select the new Term
			newTerm := getSelection().getContainer().(*Pane)
			newTerm.selected = true
			// newTerm.softRefresh()

			newTerm.vterm.RefreshCursor()
			root.refreshRenderRect()
		}
	case keypress.MouseUp:
		if mouseDownPath != nil { // end resize
			lastPathIdx := mouseDownPath[len(mouseDownPath)-1]

			parent, _ := mouseDownPath.getParent()
			first := mouseDownPath.getContainer()
			second := parent.elements[lastPathIdx+1].contents

			firstRec := first.getRenderRect()
			secondRec := second.getRenderRect()

			var combinedSize int
			if parent.verticallyStacked {
				combinedSize = firstRec.h + secondRec.h
			} else {
				combinedSize = firstRec.w + secondRec.w
			}

			var wantedRelativeBorderPos int
			if parent.verticallyStacked {
				wantedRelativeBorderPos = ev.Y - firstRec.y
			} else {
				wantedRelativeBorderPos = ev.X - firstRec.x
			}

			wantedBorderRatio := float32(wantedRelativeBorderPos) / float32(combinedSize)
			totalProportion := parent.elements[lastPathIdx].size + parent.elements[lastPathIdx+1].size

			parent.elements[lastPathIdx].size = wantedBorderRatio * totalProportion
			parent.elements[lastPathIdx+1].size = (1 - wantedBorderRatio) * totalProportion

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
