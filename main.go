package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/aaronduino/i3-tmux/keypress"
	"github.com/aaronduino/i3-tmux/render"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

var termW, termH int

var renderer *render.Renderer

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

	keypress.Listen(func(name string, data []byte) {
		// fmt.Println(name, data)

		if showingRealSelection {
			// enable artificial selection again
			fmt.Print("\033[?1000h\033[?1015h")

			renderer.Resume <- true

			showingRealSelection = false
		} else if resizeMode {
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
				path := handleMouseDown(data)
				if path != nil { // start resize
					mouseDownPath = path
				} else { // start selection
					code := string(data[3:])
					code = strings.TrimSuffix(code, "M") // NOTE: are there other codes we are forgetting about?
					pieces := strings.Split(code, ";")

					startSelectionX, _ = strconv.Atoi(pieces[1])
					startSelectionY, _ = strconv.Atoi(pieces[2])
					startSelectionX--
					startSelectionY--
				}
			case "Mouse Up":
				if mouseDownPath != nil { // end resize
					code := string(data[5:])
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
				} else { // end selection
					code := string(data[3:])
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

					selectionContent := ""

					if startSelectionY == endY {
						if startSelectionX == endX {
							return
						}

						// reverse "backwards" selections
						if startSelectionX > endX {
							tmp := startSelectionX
							startSelectionX = endX
							endX = tmp
						}

						for i := startSelectionX; i <= endX; i++ {
							selectionContent += string(renderer.GetRune(i, startSelectionY))
						}

						selectionContent = strings.TrimRight(selectionContent, " ")
					} else {
						// reverse "backwards" selections
						if startSelectionY > endY {
							tmpX := startSelectionX
							tmpY := startSelectionY
							startSelectionX = endX
							startSelectionY = endY
							endX = tmpX
							endY = tmpY
						}

						// first line
						for i := startSelectionX; i < r.x+r.w; i++ {
							selectionContent += string(renderer.GetRune(i, startSelectionY))
						}
						selectionContent = strings.TrimRight(selectionContent, " ")
						selectionContent += "\n"
						// lines in between
						for y := startSelectionY + 1; y < endY; y++ {
							for x := r.x; x < r.x+r.w; x++ {
								selectionContent += string(renderer.GetRune(x, y))
							}
						}
						selectionContent = strings.TrimRight(selectionContent, " ")
						selectionContent += "\n"
						// last line
						for i := r.x; i < endX; i++ {
							selectionContent += string(renderer.GetRune(i, endY))
						}
						selectionContent = strings.TrimRight(selectionContent, " ")
						selectionContent += "\n"
					}

					renderer.Pause <- true

					fmt.Print("\033[2J")   // clear screen
					fmt.Print("\033[0;0H") // move cursor to top left
					fmt.Print("\033[0m")   // reset styling

					// print out what we previously selected
					fmt.Print(selectionContent)

					// make real selection possible
					fmt.Print("\033[?1000l\033[?1015l")

					showingRealSelection = true
				}
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

					if config.statusBar {
						debug(root.serialize())
					}
				} else {
					t := getSelection().getContainer().(*Pane)

					t.shell.handleStdin(string(data))
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

var showingRealSelection bool

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
