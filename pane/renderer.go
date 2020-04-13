package pane

import (
	"github.com/aaronjanse/3mux/vterm"
	"github.com/rthornton128/goncurses"
)

// renderer implements vterm.Renderer
type renderer struct {
	gcWin *goncurses.Window
}

func (r *renderer) FreezeWithError(err error) {
	r.gcWin.Clear()
	r.gcWin.Print(err.Error())
}

func (r *renderer) SetChar(ch vterm.Char, x, y int) {
	r.gcWin.Move(y, x)
	r.gcWin.Print(string(ch.Rune))
	r.gcWin.Refresh()
}
