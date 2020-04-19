package pane

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/vterm"
	"github.com/aaronjanse/3mux/wm"
	"github.com/kr/pty"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	born bool

	ptmx  *os.File
	cmd   *exec.Cmd
	vterm *vterm.VTerm

	selected   bool
	renderRect wm.Rect
	renderer   *render.Renderer

	searchMode            bool
	searchText            string
	searchPos             int
	searchBackupScrollPos int
	searchDidShiftUp      bool
	searchResultsMode     bool
	searchDirection       SearchDirection

	Dead    bool
	OnDeath func(error)
}

func NewPane(renderer *render.Renderer) wm.Node {
	return newTerm(renderer)
}

func newTerm(renderer *render.Renderer) *Pane {
	shellPath, err := getShellPath()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(shellPath)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color") // FIXME we should decide whether we want 256color in $TERM
	t := &Pane{
		born:     false,
		renderer: renderer,
		cmd:      cmd,
	}

	ptmx, err := pty.Start(t.cmd)
	if err != nil {
		panic(err)
	}
	t.ptmx = ptmx

	parentSetCursor := func(x, y int) {
		if t.selected {
			renderer.SetCursor(x+t.renderRect.X, y+t.renderRect.Y)
		}
	}

	t.vterm = vterm.NewVTerm(renderer, parentSetCursor)

	return t
}

func (t *Pane) SetRenderRect(fullscreen bool, x, y, w, h int) {
	t.renderRect = wm.Rect{X: x, Y: y, W: w, H: h}

	if !t.born {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Dead = true
					t.OnDeath(fmt.Errorf("%s\n%s",
						r.(error).Error(),
						debug.Stack(),
					))
				}
			}()

			t.vterm.ProcessStdout(bufio.NewReader(t.ptmx))

			t.Dead = true
			t.OnDeath(nil)
		}()
		t.born = true
	}

	if !t.vterm.IsPaused {
		t.vterm.Reshape(x, y, w, h)
		t.vterm.RedrawWindow()
	}

	t.resizeShell(w, h)

	t.softRefresh()
}

func (t *Pane) ScrollDown() {
	t.vterm.ScrollbackDown()
}

func (t *Pane) ScrollUp() {
	t.vterm.ScrollbackUp()
}

func (t *Pane) IsDead() bool {
	return t.Dead
}

func (t *Pane) SetDeathHandler(onDeath func(error)) {
	t.OnDeath = onDeath
}

func (t *Pane) UpdateSelection(selected bool) {
	t.selected = selected
	if selected {
		t.vterm.RefreshCursor()
	}
}

func (t *Pane) HandleStdin(in string) {
	if t.searchMode && t.searchResultsMode {
		switch in[0] { // FIXME ignores extra chars
		case 'n': // next
			t.searchDirection = SearchDown
			t.searchPos--
			if t.searchPos < 0 {
				t.searchPos = 0
			}
			t.doSearch()
		case 'N': // prev
			t.searchDirection = SearchUp
			t.searchPos++
			max := len(t.vterm.Scrollback) + len(t.vterm.Screen) - 1
			if t.searchPos > max {
				t.searchPos = max
			}
			t.doSearch()
		case '/':
			t.searchResultsMode = false
			t.displayStatusText(t.searchText)
		case 127:
			fallthrough
		case 8:
			t.searchResultsMode = false
			t.searchText = t.searchText[:len(t.searchText)-1]
			t.displayStatusText(t.searchText)
		case 3:
			fallthrough
		case 4:
			fallthrough
		case 13:
			fallthrough
		case 10: // enter
			t.ToggleSearch()
			t.vterm.ScrollbackPos = t.searchPos - len(t.vterm.Screen) + t.renderRect.H/2
			t.vterm.RedrawWindow()
		}
	} else if t.searchMode {
		for _, c := range in {
			if c == 3 || c == 4 || c == 27 {
				t.ToggleSearch()
				return
			} else if c == 8 || c == 127 { // backspace
				if len(t.searchText) > 0 {
					t.searchText = t.searchText[:len(t.searchText)-1]
				}
			} else if c == 10 || c == 13 {
				if len(t.searchText) == 0 {
					t.ToggleSearch()
					return
				} else {
					t.searchResultsMode = true
					return // FIXME ignores extra chars
				}
			} else {
				t.searchText += string(c)
			}
		}
		t.searchPos = 0
		t.doSearch()
		t.displayStatusText(t.searchText)
	} else {
		t.vterm.ScrollbackReset()
		_, err := t.ptmx.Write([]byte(in))
		if err != nil {
			// fatalShutdownNow("writing to shell stdin: " + err.Error()) // FIXME
			panic(err)
		}
		t.vterm.RefreshCursor()
	}
}

func (t *Pane) Kill() {
	t.vterm.Kill()
	// FIXME: handle error
	t.ptmx.Close()
	// FIXME: handle error
	t.cmd.Process.Kill()
}

func (t *Pane) SetPaused(pause bool) {
	t.vterm.ChangePause <- pause
}

func (t *Pane) Serialize() string {
	out := fmt.Sprintf("Term[%d,%d %dx%d]", t.renderRect.X, t.renderRect.Y, t.renderRect.W, t.renderRect.H)
	if t.selected {
		return out + "*"
	}
	return out
}

func (t *Pane) simplify() {}

func (t *Pane) Resize(w, h int) {
	t.SetRenderRect(false, t.renderRect.X, t.renderRect.Y, w, h)
}

func (t *Pane) resizeShell(w, h int) {
	err := pty.Setsize(t.ptmx, &pty.Winsize{
		Rows: uint16(h), Cols: uint16(w),
		X: 16 * uint16(w), Y: 16 * uint16(h),
	})
	if err != nil {
		panic(err)
	}
}

func (t *Pane) GetRenderRect() wm.Rect {
	return t.renderRect
}

func (t *Pane) softRefresh() {
	// only selected Panes get the special highlight color
	if t.selected {
		// drawSelectionBorder(t.renderRect) // FIXME
	}
}
