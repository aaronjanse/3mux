package render

import (
	"fmt"
	"strings"
	"sync"
)

// Renderer is our simplified implemention of ncurses
type Renderer struct {
	w, h int

	currentScreen [][]Char
	pendingScreen [][]Char

	drawingCursor Cursor
	cursorMutex   *sync.Mutex
	writingMutex  *sync.Mutex

	// RenderQueue is how requests to change the framebuffer are made
	RenderQueue chan PositionedChar
}

// A PositionedChar is a char with a specific location on the screen
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
		currentScreen: [][]Char{},
		pendingScreen: [][]Char{},
		cursorMutex:   &sync.Mutex{},
		writingMutex:  &sync.Mutex{},
		RenderQueue:   make(chan PositionedChar, 10000000),
	}
}

// Resize changes the size of the framebuffer to match the host terminal size
func (r *Renderer) Resize(w, h int) {
	for y := 0; y <= h; y++ {
		if y >= len(r.pendingScreen) {
			r.pendingScreen = append(r.pendingScreen, []Char{})
		}

		for x := 0; x <= w; x++ {
			if x >= len(r.pendingScreen[y]) {
				r.pendingScreen[y] = append(r.pendingScreen[y], Char{Rune: ' ', Style: Style{}})
			}
		}
	}

	for y := 0; y <= h; y++ {
		if y >= len(r.currentScreen) {
			r.currentScreen = append(r.currentScreen, []Char{})
		}

		for x := 0; x <= w; x++ {
			if x >= len(r.currentScreen[y]) {
				r.currentScreen[y] = append(r.currentScreen[y], Char{Rune: ' ', Style: Style{}})
			}
		}
	}

	r.w = w
	r.h = h
}

func (r *Renderer) handleCh(ch PositionedChar) {
	if ch.Rune == 0 {
		ch.Rune = ' '
	}

	r.pendingScreen[ch.Y][ch.X] = Char{
		Rune:  ch.Rune,
		Style: ch.Cursor.Style,
	}
}

// ListenToQueue is a blocking function that processes data sent to the RenderQueue
func (r *Renderer) ListenToQueue() {
	for {
		ch, ok := <-r.RenderQueue
		if !ok {
			fmt.Println("Exiting scheduler")
			return
		}

		r.handleCh(ch)

	drainingLoop:
		for {
			select {
			case ch, ok := <-r.RenderQueue:
				if !ok {
					fmt.Println("Exiting scheduler")
					return
				}
				r.handleCh(ch)
			default:
				// r.cursorMutex.Lock()
				originalCursor := r.drawingCursor

				fmt.Print("\033[?25l") // hide cursor

				var diff strings.Builder
				for y := 0; y < r.h; y++ {
					for x := 0; x < r.w; x++ {
						current := r.currentScreen[y][x]
						pending := r.pendingScreen[y][x]
						if current != pending {
							r.currentScreen[y][x] = r.pendingScreen[y][x]

							newCursor := Cursor{
								X: x, Y: y, Style: pending.Style,
							}
							delta := deltaMarkup(r.drawingCursor, newCursor)
							diff.WriteString(delta)
							diff.WriteString(string(pending.Rune))
							newCursor.X++
							r.drawingCursor = newCursor
						}
					}
				}

				fmt.Print(diff.String())

				delta := deltaMarkup(originalCursor, r.drawingCursor)
				fmt.Print(delta)
				r.drawingCursor = originalCursor
				fmt.Print("\033[?25h") // show cursor

				// r.cursorMutex.Unlock()

				break drainingLoop
			}
		}
	}
}

// Refresh updates the screen to match the framebuffer
func (r *Renderer) Refresh() {
	// r.writingMutex.Lock()
	// pendingCopy := r.pendingScreen
	// r.writingMutex.Unlock()

}

// SetCursor sets the position of the physical cursor
func (r *Renderer) SetCursor(x, y int) {
	// r.cursorMutex.Lock()

	// newCursor := Cursor{
	// 	X: x, Y: y, Style: r.drawingCursor.Style,
	// }
	// delta := deltaMarkup(r.drawingCursor, newCursor)
	// fmt.Print(delta)
	// r.drawingCursor = newCursor

	// r.cursorMutex.Unlock()
}

// Debug prints the given text to the status bar
func (r *Renderer) Debug(s string) {
	// r.cursorMutex.Lock()

	// newCursor := Cursor{
	// 	X: 0, Y: r.h - 1, Style: Style{},
	// }
	// fmt.Print(deltaMarkup(r.drawingCursor, newCursor))
	// fmt.Print(s)
	// newCursor.X += len(s)
	// fmt.Print(deltaMarkup(newCursor, r.drawingCursor))

	// r.cursorMutex.Unlock()
}
