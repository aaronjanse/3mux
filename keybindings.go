package main

import (
	"fmt"
	"log"

	term "github.com/nsf/termbox-go"
)

// ModKey is a special modifier key such as Ctrl or Alt
type ModKey int

const (
	modNone ModKey = iota
	modAlt
	modCtrl
)

// ArrowKey is used to represent the up, down, left, and right arrow keys on a keyboard
type ArrowKey int

const (
	_ ArrowKey = iota
	keyArrowUp
	keyArrowDown
	keyArrowLeft
	keyArrowRight
)

// Config stores all user configuration values
type Config struct {
	bindings map[string]func()
}

var config = Config{
	bindings: map[string]func(){
		"Alt+n":           newWindow,
		"Alt+Shift+Up":    func() { moveWindow(Up) },
		"Alt+Shift+Down":  func() { moveWindow(Down) },
		"Alt+Shift+Left":  func() { moveWindow(Left) },
		"Alt+Shift+Right": func() { moveWindow(Right) },
		"Alt+Up":          func() { moveSelection(Up) },
		"Alt+Down":        func() { moveSelection(Down) },
		"Alt+Left":        func() { moveSelection(Left) },
		"Alt+Right":       func() { moveSelection(Right) },
		"Alt+Shift+Q":     killWindow,
	},
}

// listens for keypresses until it detects an escape key sequence
func listenForKeypresses() {
	err := term.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	// refreshEverything()

	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			if ev.Key == term.KeyPgdn || ev.Key == term.KeyCtrlBackslash {
				root.kill()
				close(globalCharAggregate)
				return
			}

			// fmt.Println(ev.Key, ev.Mod, ev.Ch)

			code := encodeKeypress(ev)
			if code == "" {
				continue
			}

			// fmt.Print(code)

			handleKeyCode(code)
			root.simplify()
			refreshEverything() // FIXME: inefficient; rendering should be updated by wm-ops

			// fmt.Print(ansi.MoveTo(0, termH-1) + ansi.EraseToEOL() + root.serialize())
		case term.EventError:
			panic(ev.Err)
		}
	}
}

func handleKeyCode(code string) {
	if f, ok := config.bindings[code]; ok {
		f()
	} else {
		t := getSelection().getContainer().(*Term)
		t.handleStdin(code)
	}
}

var escSeqBuffer = "" // buffer for escape sequences
func encodeKeypress(ev term.Event) string {
	r := ev.Ch

	if escSeqBuffer != "" {
		if len(escSeqBuffer) > 8 {
			escSeqBuffer = ""
		} else {
			escSeqBuffer += string(r)
		}

		// fmt.Printf("%q\n", escSeqBuffer)

		matched := true
		out := ""
		switch escSeqBuffer {
		case "\x00\x00":
			escSeqBuffer = "\x00"

		case "\x00n":
			out = "Alt+n"
		case "\x00v":
			out = "Alt+v"
		case "\x00h":
			out = "Alt+h"
		case "\x00Q":
			out = "Alt+Shift+Q"

		case "\x00[1;3A":
			out = "Alt+Up"
		case "\x00[1;3B":
			out = "Alt+Down"
		case "\x00[1;3C":
			out = "Alt+Right"
		case "\x00[1;3D":
			out = "Alt+Left"

		case "\x00[1;4A":
			out = "Alt+Shift+Up"
		case "\x00[1;4B":
			out = "Alt+Shift+Down"
		case "\x00[1;4C":
			out = "Alt+Shift+Right"
		case "\x00[1;4D":
			out = "Alt+Shift+Left"

		case "\x00[1;5A":
			out = "Ctrl+Up"
		case "\x00[1;5B":
			out = "Ctrl+Down"
		case "\x00[1;5C":
			out = "Ctrl+Right"
		case "\x00[1;5D":
			out = "Ctrl+Left"

		default:
			matched = false
		}

		if matched {
			escSeqBuffer = ""
			return out
		}
	} else if r == 0 {
		// Ctrl+[key]
		switch ev.Key {
		case term.KeyCtrlN:
			return "Ctrl+n"
		case term.KeyCtrlV:
			return "Ctrl+v"
		case term.KeyCtrlH:
			return "Ctrl+h"

		case term.KeyCtrlC:
			return "\003"
		case term.KeyCtrlD:
			return "\004"

		case term.KeyEnter:
			return "\n"

		case term.KeyBackspace2:
			return "\010"
		case term.KeyDelete:
			return "\177"

		case term.KeyArrowUp:
			return "\x1b[A"
		case term.KeyArrowDown:
			return "\x1b[B"
		case term.KeyArrowRight:
			return "\x1b[C"
		case term.KeyArrowLeft:
			return "\x1b[D"

		default:
			if ev.Key >= 32 && ev.Key <= 125 {
				return fmt.Sprintf("%c", int(ev.Key))
			}

			if r == '\x00' {
				escSeqBuffer = "\x00"
			}
		}
	}

	if escSeqBuffer != "" {
		return ""
	}

	return string(r)
}
