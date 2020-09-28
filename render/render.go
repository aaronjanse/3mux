package render

import (
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
)

// Renderer is our simplified implemention of ncurses
type Renderer struct {
	w, h int

	writingMutex  *sync.Mutex
	pendingScreen [][]ecma48.StyledChar
	currentScreen [][]ecma48.StyledChar

	highlights [][]bool

	drawingCursor ecma48.Cursor
	restingCursor ecma48.Cursor

	Pause  chan bool
	Resume chan bool

	DemoText string

	OutFd int
}

// NewRenderer returns an initialized Renderer
func NewRenderer(out int) *Renderer {
	return &Renderer{
		writingMutex:  &sync.Mutex{},
		currentScreen: [][]ecma48.StyledChar{},
		pendingScreen: [][]ecma48.StyledChar{},
		Pause:         make(chan bool),
		Resume:        make(chan bool),
		OutFd:         out,
	}
}

func (r *Renderer) Write(data []byte) {
	if r.OutFd != -1 {
		// log.Printf("%+q\n", data)
		syscall.Write(r.OutFd, data)
	}
}

// Resize changes the size of the framebuffers to match the host terminal size
func (r *Renderer) Resize(w, h int) {
	r.pendingScreen = expandBuffer(r.pendingScreen, w, h)
	r.currentScreen = expandBuffer(r.currentScreen, w, h)

	r.w = w
	r.h = h

	r.HardRefresh()
}

func expandBuffer(buffer [][]ecma48.StyledChar, w, h int) [][]ecma48.StyledChar {
	// resize currentScreen
	for y := 0; y <= h; y++ {
		if y >= len(buffer) {
			buffer = append(buffer, []ecma48.StyledChar{})
		}

		for x := 0; x <= w; x++ {
			if x >= len(buffer[y]) {
				buffer[y] = append(buffer[y], ecma48.StyledChar{Rune: ' '})
			}
		}
	}

	return buffer
}

// HandleCh places a PositionedChar in the pending screen buffer
func (r *Renderer) HandleCh(ch ecma48.PositionedChar) {
	if ch.Y < 0 || ch.Y >= len(r.pendingScreen) {
		return
	}
	if ch.X < 0 || ch.X >= len(r.pendingScreen[ch.Y]) {
		return
	}

	r.writingMutex.Lock()
	if ch.Rune == 0 {
		ch.Rune = ' '
	}

	r.pendingScreen[ch.Y][ch.X] = ecma48.StyledChar{
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
		for {
			fullyWritten := true
			var diff strings.Builder
			for y := 0; y <= r.h; y++ {
				// some terminals truncate long stdout
				if diff.Len() > 4000 {
					fullyWritten = false
					break
				}
				for x := 0; x < r.w; x++ {
					r.writingMutex.Lock()
					current := r.currentScreen[y][x]
					pending := r.pendingScreen[y][x]
					if current != pending {
						r.currentScreen[y][x] = pending

						if !pending.PrevWide {
							newCursor := ecma48.Cursor{
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
				diffBytes := []byte(diffStr)
				r.Write(diffBytes)
			}

			if fullyWritten {
				break
			}
		}

		if r.drawingCursor != r.restingCursor {
			delta := deltaMarkup(r.drawingCursor, r.restingCursor)
			r.Write([]byte(delta))
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
	r.restingCursor = ecma48.Cursor{
		X: x, Y: y, Style: r.drawingCursor.Style,
	}
}

func (r *Renderer) UpdateOut(out int) {
	r.Pause <- true
	r.OutFd = out
	r.HardRefresh()
	r.Resume <- true
}

// HardRefresh force clears all cached chars. Used for handling terminal resize
func (r *Renderer) HardRefresh() {
	r.Write([]byte("\033[2J"))
	r.Write([]byte("\033[0m"))
	r.Write([]byte("\033[H"))
	r.drawingCursor = ecma48.Cursor{}
	for y := range r.currentScreen {
		for x := range r.currentScreen[y] {
			r.currentScreen[y][x] = ecma48.StyledChar{}
		}
	}
}
