package main

import (
	"log"
	"math/rand"
	"os/exec"

	"github.com/aaronduino/i3-tmux/vterm"
	"github.com/kr/pty"
)

func newTerm(selected bool) *Pane {
	// Create arbitrary command.
	c := exec.Command("zsh")

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		log.Fatal(err)
	}

	vtermOut := make(chan vterm.Char, 32)
	vtermOperations := make(chan vterm.Operation, 32)

	vt := vterm.NewVTerm(vtermOut, vtermOperations)

	t := &Pane{
		id:       rand.Intn(10),
		selected: selected,
		ptmx:     ptmx,
		cmd:      c,

		vterm:           vt,
		vtermOut:        vtermOut,
		vtermOperations: vtermOperations,
	}

	go (func() {
		defer func() {
			if r := recover(); r != nil {
				if r != "send on closed channel" {
					panic(r)
				}
			}
		}()

		for {
			b := make([]byte, 1)
			_, err := ptmx.Read(b)
			if err != nil {
				return
			}
			r := rune(b[0])
			if r == '\t' { // FIXME
				t.vterm.Stream <- ' '
				t.vterm.Stream <- ' '
			} else {
				t.vterm.Stream <- r
			}
		}
	})()

	go (func() {
		for {
			char := <-vtermOut
			if char.Cursor.X > t.renderRect.w-1 {
				continue
			}
			if char.Cursor.Y > t.renderRect.h-1 {
				continue
			}
			char.Cursor.X += t.renderRect.x
			char.Cursor.Y += t.renderRect.y
			globalCharAggregate <- char
		}
	})()

	go (func() {
		for {
			oper := <-vtermOperations
			text := oper.Serialize(t.renderRect.x, t.renderRect.y, t.renderRect.w, t.renderRect.h, globalCursor)
			globalRawAggregate <- text
		}
	})()

	go vt.ProcessStream()

	return t
}

func (t *Pane) kill() {
	t.vterm.StopBlinker()

	close(t.vtermIn)

	err := t.ptmx.Close()
	if err != nil {
		log.Fatal("failed to close ptmx", err)
	}

	err = t.cmd.Process.Kill()
	if err != nil {
		log.Fatal("failed to kill term process", err)
	}
}

func (t *Pane) handleStdin(text string) {
	_, err := t.ptmx.Write([]byte(text))
	if err != nil {
		log.Fatal(err)
	}
}
