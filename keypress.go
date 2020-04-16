package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/aaronjanse/3mux/ecma48"
	"golang.org/x/crypto/ssh/terminal"
)

var oldState *terminal.State

// Shutdown cleans up the terminal state
func Shutdown() {
	terminal.Restore(0, oldState)

	fmt.Print("\x1b[?1002l")
	fmt.Print("\x1b[?1006l")
	fmt.Print("\x1b[?1049l")
}

func humanify(r rune) string {
	switch r {
	case '\n', '\r':
		return "Enter"
	default:
		return string(r)
	}
}

// Listen is a blocking function that indefinitely listens for keypresses.
// When it detects a keypress, it passes on to the callback a human-readable interpretation of the event (e.g. Alt+Shift+Up) along with the raw string of text received by the terminal.
func Listen(callback func(human string, obj ecma48.Output)) {
	var err error
	oldState, err = terminal.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\x1b[?1049h")
	fmt.Print("\x1b[?1006h")
	fmt.Print("\x1b[?1002h")
	fmt.Print("\x1b[?1l")

	defer Shutdown()

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			w, h, _ := GetTermSize()
			resize(w, h)
		}
	}()

	stdin := make(chan ecma48.Output, 64)

	parser := ecma48.NewParser(true)

	go parser.Parse(bufio.NewReader(os.Stdin), stdin)

	for {
		next := <-stdin

		humanCode := ""
		switch x := next.Parsed.(type) {
		case ecma48.CtrlChar:
			if x.Char == 'Q' {
				close(stdin)
				os.Stdin.Close()
				return
			}
			humanCode = fmt.Sprintf("Ctrl+%s", humanify(x.Char))
		case ecma48.AltChar:
			humanCode = fmt.Sprintf("Alt+%s", humanify(x.Char))
		case ecma48.AltShiftChar:
			humanCode = fmt.Sprintf("Alt+Shift+%s", humanify(x.Char))
		case ecma48.CursorMovement:
			if x.Ctrl {
				humanCode += "Ctrl+"
			}
			if x.Alt {
				humanCode += "Alt+"
			}
			if x.Shift {
				humanCode += "Shift+"
			}
			switch x.Direction {
			case ecma48.Up:
				humanCode += "Up"
			case ecma48.Down:
				humanCode += "Down"
			case ecma48.Left:
				humanCode += "Left"
			case ecma48.Right:
				humanCode += "Right"
			}
		}

		callback(humanCode, next)
	}
}

// GetTermSize returns the terminal dimensions w, h, err
func GetTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}
