package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"

	mathRand "math/rand"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/wm"
)

func main() {
	log.SetOutput(ioutil.Discard)

	for i := 0; i < 8; i++ {
		go fuzzECMA48()
	}

	for i := 0; i < 4; i++ {
		go fuzzWM()
	}

	for i := 0; i < 12; i++ {
		go fuzzVTerm()
	}

	wait := make(chan bool)
	<-wait
}

func fuzzVTerm() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("=== VTerm Failed ===")
			panic(r)
		}
	}()

	r := &FakeRenderer{}
	p := pane.NewPane(r)
	p.SetDeathHandler(func(err error) {
		panic(err)
	})
	p.SetRenderRect(false, 0, 0, 100, 100)

	random := bufio.NewReader(rand.Reader)
	out := make(chan ecma48.Output)

	parser := ecma48.NewParser(false)
	go parser.Parse(random, out)

	for obj := range out {
		p.HandleStdin(obj)
	}
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
	}
}

func fuzzWM() {
	var pastStates []string
	var pastFuncNames []string

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

	r := &FakeRenderer{}
	for {
		var stop bool
		u := wm.NewUniverse(r, func(err error) {
			stop = true
		}, wm.Rect{W: 100, H: 100}, newFakePane)
		pastStates = []string{}
		pastFuncNames = []string{}

		for count := 0; count < 32; count++ {
			name, fn := getRandomFunc()
			pastFuncNames = append(pastFuncNames, name)
			pastStates = append(pastStates, u.Serialize())

			fn(u)

			if u.IsDead() || stop {
				break
			}
		}
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
