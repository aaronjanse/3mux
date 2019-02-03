package main

import (
	"fmt"
	"strings"

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
		if operationCode, ok := config.bindings[name]; ok {
			executeOperationCode(operationCode)
			root.simplify()
			// refreshEverything()
		} else {
			t := getSelection().getContainer().(*Term)
			t.handleStdin(string(raw))
		}

		if config.statusBar {
			debug(root.serialize())
		}
	})

	root.kill()
}

func executeOperationCode(s string) {
	sections := strings.Split(s, "(")

	funcName := sections[0]

	var parametersText string
	if len(sections) < 2 {
		parametersText = ""
	} else {
		parametersText = strings.TrimRight(sections[1], ")")
	}
	params := strings.Split(parametersText, ",")
	for idx, param := range params {
		params[idx] = strings.TrimSpace(param)
	}

	switch funcName {
	case "newWindow":
		newWindow()
	case "moveWindow":
		d := getDirectionFromString(params[0])
		moveWindow(d)
	case "moveSelection":
		d := getDirectionFromString(params[0])
		moveSelection(d)
	case "killWindow":
		killWindow()
	default:
		panic(funcName)
	}
}

func getDirectionFromString(s string) Direction {
	switch s {
	case "Up":
		return Up
	case "Down":
		return Down
	case "Left":
		return Left
	case "Right":
		return Right
	default:
		panic(fmt.Errorf("invalid direction: %v", s))
	}
}

func debug(s string) {
	for i := 0; i < termW; i++ {
		r := ' '
		if i < len(s) {
			r = rune(s[i])
		}

		globalCharAggregate <- vterm.Char{
			Rune: r,
			Cursor: cursor.Cursor{
				X: i,
				Y: termH - 1,
				Bg: cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      2,
				},
				Fg: cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      0,
				},
			},
		}
	}
}
