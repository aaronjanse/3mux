package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"

	"github.com/aaronduino/i3-tmux/vterm"
	"github.com/kr/pty"
)

// Coords is a pair of x and y coordinates
type Coords struct {
	x, y int
}

// Markup is text that may or may not contain special escape codes such as ANSI CSI Sequences
type Markup string

// A Term is the fundamental leaf unit of screen space
type Term struct {
	id int

	selected bool

	renderRect Rect

	ptmx *os.File
	cmd  *exec.Cmd

	vterm *vterm.VTerm

	vtermIn  chan<- rune
	vtermOut <-chan vterm.Char
}

func newTerm(selected bool) *Term {
	// Create arbitrary command.
	c := exec.Command("sh")

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		log.Fatal(err)
	}

	vtermIn := make(chan rune, 32)
	vtermOut := make(chan vterm.Char, 32)

	vt := vterm.NewVTerm(vtermIn, vtermOut)

	t := &Term{
		id:       rand.Intn(10),
		selected: selected,
		ptmx:     ptmx,
		cmd:      c,

		vterm:    vt,
		vtermIn:  vtermIn,
		vtermOut: vtermOut,
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
			t.vtermIn <- rune(b[0])
		}
	})()

	go (func() {
		for {
			char := <-vtermOut
			char.Cursor.X += t.renderRect.x
			char.Cursor.Y += t.renderRect.y
			globalCharAggregate <- char
		}
	})()

	go vt.ProcessStream()

	return t
}

func (t *Term) kill() {
	close(t.vtermIn)

	err := t.ptmx.Close()
	if err != nil {
		log.Fatal("TERM_CLOSE", err)
	}

	err = t.cmd.Process.Kill()
	if err != nil {
		log.Fatal("TERM_KILL", err)
	}
}

func (t *Term) handleStdin(text string) {
	_, err := t.ptmx.Write([]byte(text))
	if err != nil {
		log.Fatal(err)
	}
}
