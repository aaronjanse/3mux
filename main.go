package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/aaronduino/i3-tmux/render"

	"github.com/aaronduino/i3-tmux/keypress"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

var termW, termH int

var renderer *render.Renderer

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	termW, termH, _ = getTermSize()

	renderer = render.NewRenderer()
	renderer.Resize(termW, termH)
	go renderer.ListenToQueue()

	root = Split{
		verticallyStacked: false,
		selectionIdx:      0,
		elements: []Node{
			Node{
				size:     1,
				contents: newTerm(true),
			},
		}}
	defer root.kill()

	var h int
	if config.statusBar {
		h = termH - 1
	} else {
		h = termH
	}
	root.setRenderRect(0, 0, termW, h)

	// if config.statusBar {
	// 	debug(root.serialize())
	// }

	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	go (func() {
		for range ticker.C {
			// for _, pane := range getPanes() {
			// 	if pane.vterm.NeedsRedraw {
			// 		pane.vterm.RedrawWindow()
			// 	}
			// }
			// renderer.Refresh()

			t := getSelection().getContainer().(*Pane)
			t.vterm.RefreshCursor()
		}
	})()

	keypress.Listen(func(name string, raw []byte) {
		// fmt.Println(name, raw)

		switch name {
		case "Scroll Up":
			t := getSelection().getContainer().(*Pane)
			t.vterm.ScrollbackDown()
		case "Scroll Down":
			t := getSelection().getContainer().(*Pane)
			t.vterm.ScrollbackUp()
		default:
			if operationCode, ok := config.bindings[name]; ok {
				executeOperationCode(operationCode)
				root.simplify()

				root.refreshRenderRect()
			} else {
				t := getSelection().getContainer().(*Pane)

				t.shell.handleStdin(string(raw))
				t.vterm.RefreshCursor()
			}
		}
	})
}

func executeOperationCode(s string) {
	sections := strings.Split(s, "(")

	funcName := sections[0]

	var parametersText string
	if len(sections) < 2 {
		parametersText = ""
	} else {
		parametersText = strings.TrimRight(sections[1], ")")
	}
	params := strings.Split(parametersText, ",")
	for idx, param := range params {
		params[idx] = strings.TrimSpace(param)
	}

	switch funcName {
	case "newWindow":
		newWindow()
	case "moveWindow":
		d := getDirectionFromString(params[0])
		moveWindow(d)
	case "moveSelection":
		d := getDirectionFromString(params[0])
		moveSelection(d)
	case "killWindow":
		killWindow()
	default:
		panic(funcName)
	}
}

func getDirectionFromString(s string) Direction {
	switch s {
	case "Up":
		return Up
	case "Down":
		return Down
	case "Left":
		return Left
	case "Right":
		return Right
	default:
		panic(fmt.Errorf("invalid direction: %v", s))
	}
}

// func debug(s string) {
// 	for i := 0; i < termW; i++ {
// 		r := ' '
// 		if i < len(s) {
// 			r = rune(s[i])
// 		}

// 		globalCharAggregate <- vterm.Char{
// 			Rune: r,
// 			Cursor: cursor.Cursor{
// 				X: i,
// 				Y: termH - 1,
// 				Bg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Bright,
// 					Code:      2,
// 				},
// 				Fg: cursor.Color{
// 					ColorMode: cursor.ColorBit3Normal,
// 					Code:      0,
// 				},
// 			},
// 		}
// 	}
// }
