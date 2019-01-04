package main

import (
	"github.com/aaronduino/i3-tmux/keypress"
)

func main() {
	go render()

	go (func() {
		t := getSelection().getContainer().(*Term)
		t.vterm.StartBlinker()
	})()

	keypress.Listen(func(name string, raw []byte) {
		if f, ok := config.bindings[name]; ok {
			f()
			root.simplify()
			refreshEverything()
		} else {
			t := getSelection().getContainer().(*Term)
			t.handleStdin(string(raw))
			refreshEverything()
		}
	})
}
