package render

import (
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
)

// This is arbitrarily chosen. Once we reach two thirds of this, we switch to frame-based rendering
const queueBufferSize = 4000

// Renderer is our simplified implemention of ncurses
type Renderer struct {
	w, h int

	renderMode uint32 // 0 for queue mode, 1 for frame mode

	charQueue chan ecma48.PositionedChar

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
		renderMode:    0,
		charQueue:     make(chan ecma48.PositionedChar, queueBufferSize),
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
	if ch.Y < 0 || ch.X < 0 || ch.Y > r.h || ch.X > r.w {
		return
	}
	if ch.Rune == 0 {
		ch.Rune = ' '
	}

	if atomic.LoadUint32(&r.renderMode) == 0 { // queue mode
		r.charQueue <- ch
	} else { // frame mode
		r.writingMutex.Lock()
		r.pendingScreen[ch.Y][ch.X] = ecma48.StyledChar{
			Rune:     ch.Rune,
			IsWide:   ch.IsWide,
			PrevWide: ch.PrevWide,
			Style:    ch.Cursor.Style,
		}
		r.writingMutex.Unlock()
	}
}

// DemoKeypress is used for demos of 3mux
func (r *Renderer) DemoKeypress(str string) {

}

func (r *Renderer) Render() {
	for {
		r.RenderViaQueue()
		r.RenderViaFrames()
	}
}

// RenderViaQueue displays characters immediately after they are sent to
// HandleCh. This rendering mode has lowest latency but can "fall behind"
// when panes are outputting characters faster than the host terminal will
// accept them. Exits during high load.
func (r *Renderer) RenderViaQueue() {
	for {
		select {
		case <-r.Pause:
			<-r.Resume
		case posCh := <-r.charQueue:
			if posCh.Rune == 0 {
				r.restingCursor = ecma48.Cursor{X: posCh.X, Y: posCh.Y, Style: r.restingCursor.Style}
			} else {
				pendingCh := ecma48.StyledChar{
					Rune:     posCh.Rune,
					IsWide:   posCh.IsWide,
					PrevWide: posCh.PrevWide,
					Style:    posCh.Cursor.Style,
				}
				if r.currentScreen[posCh.Y][posCh.X] == pendingCh {
					continue
				}
				r.currentScreen[posCh.Y][posCh.X] = pendingCh

				delta := deltaMarkup(r.drawingCursor, posCh.Cursor) + string(posCh.Rune)
				r.Write([]byte(delta))
				r.drawingCursor = posCh.Cursor
				if posCh.IsWide {
					r.drawingCursor.X += 2
				} else {
					r.drawingCursor.X++
				}
			}
			delta := deltaMarkup(r.drawingCursor, r.restingCursor)
			r.Write([]byte(delta))
			r.drawingCursor = r.restingCursor

			// log.Println("# Queue length:", len(r.charQueue))
			if len(r.charQueue) > (queueBufferSize*2)/3 {
				r.writingMutex.Lock()
				atomic.StoreUint32(&r.renderMode, 1)
				for y := 0; y <= r.h; y++ {
					copy(r.pendingScreen[y], r.currentScreen[y])
				}
				for posCh := range r.charQueue {
					r.pendingScreen[posCh.Y][posCh.X] = ecma48.StyledChar{
						Rune:     posCh.Rune,
						IsWide:   posCh.IsWide,
						PrevWide: posCh.PrevWide,
						Style:    posCh.Cursor.Style,
					}
					if len(r.charQueue) == 0 {
						break
					}
				}
				r.writingMutex.Unlock()
				return
			}
		}
	}
}

// RenderViaFrames sets a framerate at which it writes all new characters.
// This is best during high load, such as dumping a file to stdout, where
// the renderer can quickly show the final display state instead of slowly
// scrolling to get there. Exits during low load.
func (r *Renderer) RenderViaFrames() {
	numEmptyFrames := 0
	const numEmptyFramesTillExit = 10
	for {
		if numEmptyFrames >= numEmptyFramesTillExit {
			atomic.StoreUint32(&r.renderMode, 0)
		}

		emptyFrame := r.RenderSingleFrame()

		if r.drawingCursor != r.restingCursor {
			delta := deltaMarkup(r.drawingCursor, r.restingCursor)
			r.Write([]byte(delta))
			r.drawingCursor = r.restingCursor
		}

		if numEmptyFrames >= numEmptyFramesTillExit {
			break
		}

		if emptyFrame {
			numEmptyFrames++
		} else {
			numEmptyFrames = 0
		}

		// this delay frees up the CPU for an arbitrary amount of time
		timer := time.NewTimer(time.Millisecond * 25)

		select {
		case <-timer.C:
			timer.Stop()
		case <-r.Pause:
			<-r.Resume
		}
	}
}

func (r *Renderer) RenderSingleFrame() bool {
	emptyFrame := true
	for {
		fullyWritten := true
		var diff strings.Builder
	outer:
		for y := 0; y <= r.h; y++ {
			for x := 0; x < r.w; x++ {
				// some terminals truncate long stdout
				// 4000 to allow some extra chars to be added before crossing 4096
				if diff.Len() > 4000 {
					fullyWritten = false
					break outer
				}
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
			emptyFrame = false
			diffBytes := []byte(diffStr)
			// log.Println("Writing frame diff of length", len(diffBytes))
			// log.Printf("Writing frame bytes: %q", diffBytes)
			r.Write(diffBytes)
		}

		if fullyWritten {
			break
		}
	}
	return emptyFrame
}

// SetCursor sets the position of the physical cursor
func (r *Renderer) SetCursor(x, y int) {
	if atomic.LoadUint32(&r.renderMode) == 0 { // queue mode
		r.charQueue <- ecma48.PositionedChar{Cursor: ecma48.Cursor{X: x, Y: y}}
	} else { // frame mode
		r.restingCursor = ecma48.Cursor{X: x, Y: y, Style: r.restingCursor.Style}
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
	r.Write([]byte("\033[0m"))
	r.Write([]byte("\033[2J"))
	r.Write([]byte("\033[H"))
	r.drawingCursor = ecma48.Cursor{}
	r.currentScreen = expandBuffer([][]ecma48.StyledChar{}, r.w, r.h)

	r.RenderSingleFrame()
}
