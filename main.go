package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	runtimeDebug "runtime/debug"
	"runtime/pprof"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/render"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

var termW, termH int

var renderer *render.Renderer

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var writeLogs = flag.Bool("log", false, "write logs to ./logs.txt")

func main() {
	defer func() {
		if r := recover(); r != nil {
			fatalShutdownNow("main.go\n" + r.(error).Error())
		}
	}()

	flag.Parse()

	// setup logging
	if *writeLogs {
		f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// setup cpu profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var err error
	termW, termH, err = GetTermSize()
	if err != nil {
		log.Fatalf("While getting terminal size: %s", err.Error())
	}

	renderer = render.NewRenderer()
	go renderer.ListenToQueue()

	root = Universe{
		workspaces: []*Workspace{
			&Workspace{
				contents: &Split{
					verticallyStacked: false,
					selectionIdx:      0,
					elements: []Node{
						Node{
							size:     1,
							contents: newTerm(true),
						},
					}},
				doFullscreen: false,
			},
		},
		selectionIdx: 0,
	}

	defer root.kill()

	resize(termW, termH)

	if config.statusBar {
		debug(root.serialize())
	}

	if demoMode {
		go doDemo()
	}

	Listen(handleInput)
}

func resize(w, h int) {
	termW = w
	termH = h

	renderer.Resize(w, h)

	var wmH int
	if config.statusBar {
		wmH = h - 1
	} else {
		wmH = h
	}
	root.setRenderRect(0, 0, w, wmH)

	renderer.HardRefresh()
}

func shutdownNow() {
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	Shutdown()
	os.Exit(0)
}

func fatalShutdownNow(where string) {
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	Shutdown()
	fmt.Println("Error during:", where)
	fmt.Println("Tiling state:", root.serialize())
	fmt.Println(string(runtimeDebug.Stack()))
	fmt.Println()
	fmt.Println("Please submit a bug report with this stack trace to https://github.com/aaronjanse/3mux/issues")
	os.Exit(0)
}

var resizeMode bool

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
					Bg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Bright,
						Code:      2,
					},
					Fg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		renderer.HandleCh(ch)
	}

	if resizeMode {
		resizeText := "RESIZE"

		for i, r := range resizeText {
			ch := render.PositionedChar{
				Rune: r,
				Cursor: render.Cursor{
					X: termW - len(resizeText) + i,
					Y: termH - 1,
					Style: render.Style{
						Bg: ecma48.Color{
							ColorMode: ecma48.ColorBit3Bright,
							Code:      3,
						},
						Fg: ecma48.Color{
							ColorMode: ecma48.ColorBit3Normal,
							Code:      0,
						},
					},
				},
			}
			renderer.HandleCh(ch)
		}
	}
}
