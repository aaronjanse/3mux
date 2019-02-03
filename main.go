package main

import (
	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/keypress"
	"github.com/aaronduino/i3-tmux/vterm"
)

func main() {
	go render()

	t := getSelection().getContainer().(*Term)
	t.vterm.StartBlinker()

	refreshEverything()

	keypress.Listen(func(name string, raw []byte) {
		// fmt.Println(name)
		if f, ok := config.bindings[name]; ok {
			f()
			root.simplify()
			refreshEverything()
			// debug(root.serialize())
		} else {
			t := getSelection().getContainer().(*Term)
			t.handleStdin(string(raw))
		}
	})

	root.kill()
}

func debug(s string) {
	for i, r := range []rune(s) {
		globalCharAggregate <- vterm.Char{
			Rune: r,
			Cursor: cursor.Cursor{
				X: i,
				Y: termH - 1,
			},
		}
	}
}
