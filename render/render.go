package render

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Renderer is our simplified implemention of ncurses
type Renderer struct {
	w, h int

	writingMutex  *sync.Mutex
	pendingScreen [][]Char
	currentScreen [][]Char

	highlights [][]bool

	drawingCursor Cursor
	restingCursor Cursor

	Pause  chan bool
	Resume chan bool

	DemoText string
}

// A PositionedChar is a Char with a specific location on the screen
type PositionedChar struct {
	Rune     rune
	IsWide   bool
	PrevWide bool
	Cursor
}

// A Char is a rune with a visual style associated with it
type Char struct {
	Rune     rune
	IsWide   bool
	PrevWide bool
	Style
}

// NewRenderer returns an initialized Renderer
func NewRenderer() *Renderer {
	return &Renderer{
		writingMutex:  &sync.Mutex{},
		currentScreen: [][]Char{},
		pendingScreen: [][]Char{},
		Pause:         make(chan bool),
		Resume:        make(chan bool),
	}
}

// Resize changes the size of the framebuffers to match the host terminal size
func (r *Renderer) Resize(w, h int) {
	r.pendingScreen = expandBuffer(r.pendingScreen, w, h)
	r.currentScreen = expandBuffer(r.currentScreen, w, h)

	r.w = w
	r.h = h
}

func expandBuffer(buffer [][]Char, w, h int) [][]Char {
	// resize currentScreen
	for y := 0; y <= h; y++ {
		if y >= len(buffer) {
			buffer = append(buffer, []Char{})
		}

		for x := 0; x <= w; x++ {
			if x >= len(buffer[y]) {
				buffer[y] = append(buffer[y], Char{Rune: ' '})
			}
		}
	}

	return buffer
}

// HandleCh places a PositionedChar in the pending screen buffer
func (r *Renderer) HandleCh(ch PositionedChar) {
	r.writingMutex.Lock()
	if ch.Rune == 0 {
		ch.Rune = ' '
	}

	r.pendingScreen[ch.Y][ch.X] = Char{
		Rune:     ch.Rune,
		IsWide:   ch.IsWide,
		PrevWide: ch.PrevWide,
		Style:    ch.Cursor.Style,
	}
	r.writingMutex.Unlock()
}

// DemoKeypress is used for demos of 3mux
func (r *Renderer) DemoKeypress(str string) {

}

// ListenToQueue is a blocking function that processes data sent to the RenderQueue
func (r *Renderer) ListenToQueue() {
	for {
		var diff strings.Builder
		for y := 0; y <= r.h; y++ {
			for x := 0; x < r.w; x++ {
				r.writingMutex.Lock()
				current := r.currentScreen[y][x]
				pending := r.pendingScreen[y][x]
				if current != pending {
					r.currentScreen[y][x] = pending

					if !pending.PrevWide {
						newCursor := Cursor{
							X: x, Y: y, Style: pending.Style,
						}

						delta := deltaMarkup(r.drawingCursor, newCursor)
						diff.WriteString(delta)
						diff.WriteString(string(pending.Rune))

						if pending.IsWide {
							newCursor.X += 2
						} else {
							newCursor.X++
						}

						r.drawingCursor = newCursor
					}
				}
				r.writingMutex.Unlock()
			}
		}

		diffStr := diff.String()
		if len(diffStr) > 0 {
			// fmt.Print("\033[?25l") // hide cursor

			fmt.Print(diffStr)
			// log.Printf("RENDER: %+q\n", diffStr)

			if len(r.DemoText) > 0 {
				var demoTextDiff strings.Builder

				demoTextLen := utf8.RuneCountInString(r.DemoText)

				for x := r.w - 2 - demoTextLen - 1; x <= r.w-2; x++ {
					for y := r.h - 5; y <= r.h-3; y++ {
						newCursor := Cursor{
							X: x, Y: y, Style: Style{
								Bg: Color{
									ColorMode: ColorBit3Bright,
									Code:      6,
								},
								Fg: Color{
									ColorMode: ColorBit3Normal,
									Code:      0,
								},
							},
						}

						delta := deltaMarkup(r.drawingCursor, newCursor)
						demoTextDiff.WriteString(delta)
						demoTextDiff.WriteString(string(' '))
						newCursor.X++
						r.drawingCursor = newCursor
					}
				}

				for i, c := range r.DemoText {
					newCursor := Cursor{
						X: r.w - 2 - demoTextLen + i, Y: r.h - 4, Style: Style{
							Bg: Color{
								ColorMode: ColorBit3Bright,
								Code:      6,
							},
							Fg: Color{
								ColorMode: ColorBit3Normal,
								Code:      0,
							},
						},
					}

					delta := deltaMarkup(r.drawingCursor, newCursor)
					demoTextDiff.WriteString(delta)
					demoTextDiff.WriteString(string(c))
					newCursor.X++
					r.drawingCursor = newCursor
				}

				fmt.Print(demoTextDiff.String())
			}

			// fmt.Print("\033[?25h") // show cursor
		}

		if r.drawingCursor != r.restingCursor {
			delta := deltaMarkup(r.drawingCursor, r.restingCursor)
			fmt.Print(delta)
			r.drawingCursor = r.restingCursor
		}

		// thr delay frees up the CPU for an arbitrary amount of time
		timer := time.NewTimer(time.Millisecond * 25)

		select {
		case <-timer.C:
			timer.Stop()
		case <-r.Pause:
			<-r.Resume
			// fmt.Print("\033[0;0H\033[0m") // reset real cursor
			// r.drawingCursor = Cursor{}    // reset virtual cursor
		}
	}
}

// SetCursor sets the position of the physical cursor
func (r *Renderer) SetCursor(x, y int) {
	r.restingCursor = Cursor{
		X: x, Y: y, Style: r.drawingCursor.Style,
	}
}

// Debug prints the given text to the status bar
func (r *Renderer) Debug(s string) {
	for i, ch := range s {
		r.HandleCh(PositionedChar{
			Rune: rune(ch),
			Cursor: Cursor{
				X: i, Y: r.h - 1,
				Style: Style{},
			}})
	}
}

// GetRune returns the rune of the currentScreen at the given coordinates
func (r *Renderer) GetRune(x, y int) rune {
	return r.currentScreen[y][x].Rune
}

// HardRefresh force clears all cached chars. Used for handling terminal resize
func (r *Renderer) HardRefresh() {
	log.Println("HARD REFRESH")
	fmt.Print("\033[2J")
	fmt.Print("\033[0m")
	fmt.Print("\033[H")
	r.drawingCursor = Cursor{}
	for y := range r.currentScreen {
		for x := range r.currentScreen[y] {
			r.currentScreen[y][x] = Char{Rune: ' '}
		}
	}
}
