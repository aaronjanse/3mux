package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/wm"
	"golang.org/x/crypto/ssh/terminal"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var writeLogs = flag.Bool("log", false, "write logs to ./logs.txt")

func main() {
	if os.Getenv("THREEMUX") == "1" {
		fmt.Println("Refusing to run 3mux inside itself.")
		fmt.Println("If you want to do it anyway, `unset THREEMUX`.")
		return
	}

	shutdown := make(chan error)
	stateBeforeInput := ""

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

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r.(error))
			fmt.Println(string(debug.Stack()))
			fmt.Println()
			fmt.Printf("BEFORE `%s`", stateBeforeInput)
			fmt.Println()
			fmt.Println("Please report this to https://github.com/aaronjanse/3mux/issues.")
		}
	}()

	config := loadOrGenerateConfig()

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Restore(0, oldState)

	termW, termH, err := GetTermSize()
	if err != nil {
		log.Fatalf("While getting terminal size: %s", err.Error())
	}

	renderer := render.NewRenderer()
	renderer.Resize(termW, termH)
	go renderer.ListenToQueue()

	u := wm.NewUniverse(renderer, config.generalSettings.EnableHelpBar, func(err error) {
		go func() {
			if err != nil {
				shutdown <- fmt.Errorf("%s\n%s", err.Error(), debug.Stack())
			} else {
				shutdown <- nil
			}
		}()
	}, wm.Rect{X: 0, Y: 0, W: termW, H: termH}, pane.NewPane)
	defer u.Kill()

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			w, h, _ := GetTermSize()
			renderer.Resize(w, h)
			u.SetRenderRect(0, 0, w, h)
		}
	}()

	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")
	fmt.Print("\x1b[?1006h")
	defer fmt.Print("\x1b[?1006l")
	fmt.Print("\x1b[?1002h")
	defer fmt.Print("\x1b[?1002l")

	fmt.Print("\x1b[?1l")

	stdin := make(chan ecma48.Output, 64)
	defer close(stdin)

	parser := ecma48.NewParser(true)

	go parser.Parse(bufio.NewReader(os.Stdin), stdin)

	for {
		select {
		case next := <-stdin:
			stateBeforeInput = u.Serialize()

			human := humanify(next)

			if human == "Ctrl+Q" {
				return
			}

			if seiveMouseEvents(u, human, next) {
				break
			}
			if seiveConfigEvents(config, u, human) {
				break
			}
			// if we didn't find anything special, just pass the raw data to
			// the selected terminal
			u.HandleStdin(next)
		case err := <-shutdown:
			if err != nil {
				panic(err)
			} else {
				return
			}
		}
	}
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

// GetTermSize returns the terminal dimensions w, h, err
func GetTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}
