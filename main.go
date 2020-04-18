package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/wm"
)

var termW, termH int

var renderer *render.Renderer

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var writeLogs = flag.Bool("log", false, "write logs to ./logs.txt")

func main() {
	shutdown := make(chan bool)

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r.(error))
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

	root := wm.NewUniverse(renderer, func(err error) {
		shutdown <- true
	}, wm.Rect{X: 0, Y: 0, W: termW, H: termH}, pane.NewPane)

	defer root.Kill()

	if demoMode {
		go doDemo()
	}

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			w, h, _ := GetTermSize()
			root.SetRenderRect(0, 0, w, h)
		}
	}()

	go Listen(func(human string, obj ecma48.Output) {
		handleInput(root, human, obj)
	})
	defer shutdownNow()

	<-shutdown
}

func shutdownNow() {
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	Shutdown()
	os.Exit(0)
}

var resizeMode bool

func getDirectionFromString(s string) wm.Direction {
	switch s {
	case "Up":
		return wm.Up
	case "Down":
		return wm.Down
	case "Left":
		return wm.Left
	case "Right":
		return wm.Right
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
