package pane

import (
	"C"

	"sync"

	"github.com/aaronjanse/3mux/vterm"
	gc "github.com/rthornton128/goncurses"
)

// renderer implements vterm.Renderer
type renderer struct {
	gcWin *gc.Window
	lock  *sync.Mutex
}

func (r *renderer) FreezeWithError(err error) {
	r.gcWin.Clear()
	r.gcWin.Print(err.Error())
}

func (r *renderer) SetChar(ch vterm.Char, x, y int) {
	r.lock.Lock()
	ru := ch.Rune
	if ru == 0 {
		ru = ' '
	}
	// r.gcWin.MoveAddChar(y, x, gc.Char(ru))
	r.gcWin.Move(y, x)
	r.gcWin.Print(string(ru))
	r.lock.Unlock()
	// log.Println(time.Now())
}

func (r *renderer) Refresh() {
	r.lock.Lock()
	r.gcWin.Refresh()
	r.lock.Unlock()
	// log.Println(time.Now())
}

func (r *renderer) RefreshCursor() {
	r.lock.Lock()
	y, x := r.gcWin.CursorYX()
	r.gcWin.Move(y, x)
	r.gcWin.Refresh()
	r.lock.Unlock()
}

func (r *renderer) SetCursor(x, y int) {
	r.lock.Lock()
	r.gcWin.Move(y, x)
	r.gcWin.Refresh()
	r.lock.Unlock()
}
