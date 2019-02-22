package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aaronduino/i3-tmux/keypress"
	"github.com/aaronduino/i3-tmux/render"

	term "github.com/nsf/termbox-go"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

var termW, termH int

var renderer *render.Renderer

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
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

	// enable mouse reporting
	fmt.Print("\033[?1000h")
	defer fmt.Print("\033[?1000l")
	fmt.Print("\033[?1005h")
	defer fmt.Print("\033[?1005l")
	fmt.Print("\033[?1015h")
	defer fmt.Print("\033[?1015l")

	keypress.Listen(handleInput)
}

func shutdownNow() {
	term.Close()
	os.Exit(0)
}

var resizeMode bool
var mouseDownPath Path

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
