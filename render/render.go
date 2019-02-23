package render

import (
	"fmt"
	"strings"
	"sync"
	"time"
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
}

// A PositionedChar is a Char with a specific location on the screen
type PositionedChar struct {
	Rune rune
	Cursor
}

// A Char is a rune with a visual style associated with it
type Char struct {
	Rune rune
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
		Rune:  ch.Rune,
		Style: ch.Cursor.Style,
	}
	r.writingMutex.Unlock()
}

// ForceHandleCh places a PositionedChar in the pending screen buffer, ignoring cache buffering
func (r *Renderer) ForceHandleCh(ch PositionedChar) {
	r.currentScreen[ch.Y][ch.X] = Char{
		Rune: 0,
	}
	r.HandleCh(ch)
}

// ListenToQueue is a blocking function that processes data sent to the RenderQueue
func (r *Renderer) ListenToQueue() {
	fmt.Print("\033[2J") // clear screen

	for {
		fmt.Print("\033[?25l") // hide cursor

		var diff strings.Builder
		for y := 0; y < r.h; y++ {
			for x := 0; x < r.w; x++ {
				r.writingMutex.Lock()
				current := r.currentScreen[y][x]
				pending := r.pendingScreen[y][x]
				if current != pending {
					r.currentScreen[y][x] = pending

					newCursor := Cursor{
						X: x, Y: y, Style: pending.Style,
					}

					delta := deltaMarkup(r.drawingCursor, newCursor)
					diff.WriteString(delta)
					diff.WriteString(string(pending.Rune))
					newCursor.X++
					r.drawingCursor = newCursor
				}
				r.writingMutex.Unlock()
			}
		}

		fmt.Print(diff.String())

		// put the cursor back in its resting position
		delta := deltaMarkup(r.drawingCursor, r.restingCursor)
		fmt.Print(delta)
		r.drawingCursor = r.restingCursor

		fmt.Print("\033[?25h") // show cursor

		// thr delay frees up the CPU for an arbitrary amount of time
		timer := time.NewTimer(time.Millisecond * 25)

		select {
		case <-timer.C:
			timer.Stop()
		case <-r.Pause:
			<-r.Resume
			fmt.Print("\033[0;0H\033[0m") // reset real cursor
			r.drawingCursor = Cursor{}    // reset virtual cursor
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

// HardRefresh force clears all cached chars
func (r *Renderer) HardRefresh() {
	fmt.Print("\033[2J")
	for y := range r.currentScreen {
		for x := range r.currentScreen[y] {
			r.currentScreen[y][x].Rune = ' '
		}
	}
}
