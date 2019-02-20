package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/aaronduino/i3-tmux/keypress"
	"github.com/aaronduino/i3-tmux/render"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

var termW, termH int

var renderer *render.Renderer

var startTime int64

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	shutdown = make(chan bool, 20)

	// setup logging
	f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// setup cpu profiling
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

	startTime = time.Now().UnixNano()

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

	/* enable mouse reporting */

	fmt.Print("\033[?1000h")
	defer fmt.Print("\033[?1000l")

	fmt.Print("\033[?1015h")
	defer fmt.Print("\033[?1015l")

	/* listen for keypresses */

	keypress.Listen(func(name string, raw []byte) {
		// fmt.Println(name, raw)
		if resizeMode {
			switch name {
			case "Up", "Down", "Right", "Left":
				d := getDirectionFromString(name)
				resizeWindow(d, 0.1)
			case "Enter":
				resizeMode = false
			}
		} else {
			switch name {
			case "Mouse Down":
				path := handleMouseDown(raw)
				if path != nil {
					mouseDownPath = path
				}
			case "Mouse Up":
				if mouseDownPath != nil {
					code := string(raw[5:])
					parts := strings.Split(code, ";")
					x, _ := strconv.Atoi(parts[0])
					y, _ := strconv.Atoi(strings.TrimSuffix(parts[1], "M"))

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
			case "Scroll Up":
				t := getSelection().getContainer().(*Pane)
				t.vterm.ScrollbackDown()
			case "Scroll Down":
				t := getSelection().getContainer().(*Pane)
				t.vterm.ScrollbackUp()
			case "Start Selection":
				code := string(raw[3:])
				code = strings.TrimSuffix(code, "M") // NOTE: are there other codes we are forgetting about?
				pieces := strings.Split(code, ";")

				startSelectionX, _ = strconv.Atoi(pieces[1])
				startSelectionY, _ = strconv.Atoi(pieces[2])
				startSelectionX--
				startSelectionY--
			case "End Selection":
				code := string(raw[3:])
				code = strings.TrimSuffix(code, "M") // NOTE: are there other codes we are forgetting about?
				pieces := strings.Split(code, ";")

				endX, _ := strconv.Atoi(pieces[1])
				endY, _ := strconv.Atoi(pieces[2])
				endX--
				endY--

				path := findClosestBorderForCoord([]int{}, startSelectionX, startSelectionY)
				r := path.getContainer().getRenderRect()

				if endX > r.x+r.w {
					endX = r.x + r.w
				}
				if endY > r.y+r.h {
					endY = r.y + r.h
				}

				// TODO: support "backwards" selections

				if startSelectionY == endY {
					for i := startSelectionX; i <= endX; i++ {
						log.Println(i, startSelectionY)
						renderer.Highlight(i, startSelectionY)
					}
				} else {
					// highlight the first line
					for i := startSelectionX; i < r.x+r.w; i++ {
						renderer.Highlight(i, startSelectionY)
					}

					// highlight the last line
					for i := r.x; i < endX; i++ {
						renderer.Highlight(i, endY)
					}

					// highlight the lines in between
					for y := startSelectionY + 1; y < endY; y++ {
						for x := r.x; x < r.x+r.w; x++ {
							renderer.Highlight(x, y)
						}
					}
				}
			default:
				if operationCode, ok := config.bindings[name]; ok {
					executeOperationCode(operationCode)
					root.simplify()

					root.refreshRenderRect()

					if config.statusBar {
						debug(root.serialize())
					}
				} else {
					t := getSelection().getContainer().(*Pane)

					t.shell.handleStdin(string(raw))
					t.vterm.RefreshCursor()
				}
			}
		}
	})

	// <-shutdown
}

var shutdown chan bool

var resizeMode bool
var mouseDownPath Path

var startSelectionX int
var startSelectionY int

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
	case "resize":
		resizeMode = true
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

func debug(s string) {
	for i := 0; i < termW; i++ {
		r := ' '
		if i < len(s) {
			r = rune(s[i])
		}

		ch := render.PositionedChar{
			Rune: r,
			Cursor: render.Cursor{
				X: i,
				Y: termH - 1,
				Style: render.Style{
					Bg: render.Color{
						ColorMode: render.ColorBit3Bright,
						Code:      2,
					},
					Fg: render.Color{
						ColorMode: render.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		renderer.HandleCh(ch)
	}
}
