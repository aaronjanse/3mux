package pane

import (
	"C"

	"github.com/aaronjanse/3mux/vterm"
	gc "github.com/rthornton128/goncurses"
)

// renderer implements vterm.Renderer
type renderer struct {
	gcWin *gc.Window
}

func (r *renderer) FreezeWithError(err error) {
	r.gcWin.Clear()
	r.gcWin.Print(err.Error())
}

func (r *renderer) SetChar(ch vterm.Char, x, y int) {
	ru := ch.Rune
	if ru == 0 {
		ru = ' '
	}
	// r.gcWin.MoveAddChar(y, x, gc.Char(ru))
	r.gcWin.Move(y, x)
	r.gcWin.Print(string(ru))
	// log.Println(time.Now())
}

func (r *renderer) Refresh() {
	r.gcWin.Refresh()
	// log.Println(time.Now())
}
