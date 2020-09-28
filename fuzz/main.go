package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"

	mathRand "math/rand"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/vterm"
	"github.com/aaronjanse/3mux/wm"
)

var ecmaCount int
var wmCount int

func main() {
	log.SetOutput(ioutil.Discard)

	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("go run *.go wm")
		fmt.Println("go run *.go vterm")
		fmt.Println("go run *.go ecma48")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ecma48":
		for i := 0; i < 8; i++ {
			go fuzzECMA48()
		}
	case "wm":
		for i := 0; i < 2; i++ {
			go fuzzWM()
		}
	case "vterm":
		for i := 0; i < 8; i++ {
			go fuzzVTerm()
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			p := message.NewPrinter(language.English)
			p.Println()
			p.Printf("WM operations:     %d\n", wmCount)
			p.Printf("ECMA48 operations: %d\n", ecmaCount)
			p.Printf("VTerm operations:  (not recorded)\n")

			os.Exit(0)
		}
	}()

	wait := make(chan bool)
	<-wait
}

func fuzzVTerm() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("=== VTerm Failed ===")
			panic(r.(error))
		}
	}()

	r := &FakeRenderer{}
	p := pane.NewPane(r, false, "1")
	p.SetDeathHandler(func(err error) {
		panic(err)
	})
	p.SetRenderRect(false, 0, 0, 100, 100)

	random := bufio.NewReader(rand.Reader)

	vterm := vterm.NewVTerm(r, func(x, y int) {})
	vterm.ProcessStdout(bufio.NewReader(random))
}

func fuzzECMA48() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("=== ECMA48 Failed ===")
			panic(r)
		}
	}()

	random := bufio.NewReader(rand.Reader)
	out := make(chan ecma48.Output)

	p := ecma48.NewParser(false)
	go p.Parse(random, out)

	for range out {
		ecmaCount++
	}
}

func fuzzWM() {
	var pastStates []string
	var pastFuncNames []string

	r := &FakeRenderer{}
	for {
		var stop bool
		u := wm.NewUniverse(r, false, false, func(err error) {
			stop = true
		}, wm.Rect{W: 100, H: 100}, newFakePane)
		pastStates = []string{}
		pastFuncNames = []string{}

		var wg sync.WaitGroup

		for count := 0; count < 4; count++ {
			name, fn := getRandomFunc()
			pastFuncNames = append(pastFuncNames, name)
			pastStates = append(pastStates, u.Serialize())

			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("=== WM Failed ===")
						for i, state := range pastStates {
							fmt.Println("State:", state)
							fmt.Println("Func: ", pastFuncNames[i])
						}
						panic(r)
					}
				}()

				fn(u)
				wg.Done()
			}()

			if u.IsDead() || stop {
				break
			}
			wmCount++
		}

		wg.Wait()
	}
}

func getRandomFunc() (string, func(*wm.Universe)) {
	i := mathRand.Intn(len(wm.FuncNames))
	for k, v := range wm.FuncNames {
		if i == 0 {
			return k, v
		}
		i--
	}
	panic("should not be possible")
}

type FakeRenderer struct{}

func (r *FakeRenderer) HandleCh(ch ecma48.PositionedChar) {
}
func (r *FakeRenderer) SetCursor(x, y int) {
}

type FakePane struct {
	rect wm.Rect
	dead bool
}

func newFakePane(renderer ecma48.Renderer) wm.Node {
	return &FakePane{}
}
func (p *FakePane) SetRenderRect(fullscreen bool, x, y, w, h int) {
	p.rect = wm.Rect{X: x, Y: y, W: w, H: h}
}
func (p *FakePane) GetRenderRect() wm.Rect {
	return p.rect
}
func (p *FakePane) Serialize() string {
	return "FakePane"
}
func (p *FakePane) SetPaused(paused bool) {
}
func (p *FakePane) Kill() {
	p.dead = true
}
func (p *FakePane) IsDead() bool {
	return p.dead
}
func (p *FakePane) ScrollUp() {
}
func (p *FakePane) ScrollDown() {
}
func (p *FakePane) ToggleSearch() {
}
func (p *FakePane) HandleStdin(ecma48.Output) {
}
func (p *FakePane) UpdateSelection(selected bool) {
}
func (p *FakePane) SetDeathHandler(fn func(error)) {
}
