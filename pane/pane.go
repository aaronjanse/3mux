package pane

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/vterm"
	"github.com/aaronjanse/3mux/wm"
	"github.com/aaronjanse/pty"
)

// A Pane is a tiling unit representing a terminal
type Pane struct {
	born bool

	ptmx  *os.File
	cmd   *exec.Cmd
	vterm *vterm.VTerm

	selected   bool
	renderRect wm.Rect
	renderer   ecma48.Renderer

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

func NewPane(renderer ecma48.Renderer, realShell bool, sessionID string) wm.Node {
	shellPath, err := getShellPath()
	if err != nil {
		panic(err)
	}

	if !realShell {
		shellPath = "cat"
	}

	cmd := exec.Command(shellPath)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color") // FIXME we should decide whether we want 256color in $TERM
	cmd.Env = append(cmd.Env, fmt.Sprintf("THREEMUX=%s", sessionID))
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
						r.(error), debug.Stack(),
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

func (t *Pane) HandleStdin(in ecma48.Output) {
	if t.searchMode {
		t.handleSearchStdin(string(in.Raw))
	} else {
		t.vterm.ScrollbackReset()
		_, err := t.ptmx.Write(t.vterm.ProcessStdin(in))
		if err != nil {
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

	t.Dead = true
}

func (t *Pane) SetPaused(pause bool) {
	t.vterm.ChangePause <- pause
	t.vterm.IsPaused = pause
}

func (t *Pane) Serialize() string {
	out := fmt.Sprintf("Term[%d,%d %dx%d]", t.renderRect.X, t.renderRect.Y, t.renderRect.W, t.renderRect.H)
	if t.selected {
		return out + "*"
	}
	return out
}

func (t *Pane) GetRenderRect() wm.Rect {
	return t.renderRect
}
