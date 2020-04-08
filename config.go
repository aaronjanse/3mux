package main

import (
	"log"
	"strings"

	"github.com/aaronjanse/i3-tmux/keypress"
)

// Config stores all user configuration values
type Config struct {
	statusBar bool
	bindings  map[interface{}]string
}

var config = Config{
	statusBar: true,
	bindings: map[interface{}]string{
		keypress.AltChar{Char: 'N'}:  "newWindow",
		keypress.AltChar{Char: '\n'}: "newWindow",
		keypress.AltChar{Char: 'F'}:  "fullscreen",

		keypress.AltChar{Char: 'X'}: "debugSlowMode",

		keypress.AltChar{Char: '/'}: "search",

		keypress.AltShiftArrow{Direction: keypress.Up}:    "moveWindow(Up)",
		keypress.AltShiftArrow{Direction: keypress.Down}:  "moveWindow(Down)",
		keypress.AltShiftArrow{Direction: keypress.Left}:  "moveWindow(Left)",
		keypress.AltShiftArrow{Direction: keypress.Right}: "moveWindow(Right)",

		keypress.AltShiftChar{Char: 'H'}: "moveWindow(Left)",
		keypress.AltShiftChar{Char: 'J'}: "moveWindow(Down)",
		keypress.AltShiftChar{Char: 'K'}: "moveWindow(Up)",
		keypress.AltShiftChar{Char: 'L'}: "moveWindow(Right)",

		keypress.AltArrow{Direction: keypress.Up}:    "moveSelection(Up)",
		keypress.AltArrow{Direction: keypress.Down}:  "moveSelection(Down)",
		keypress.AltArrow{Direction: keypress.Left}:  "moveSelection(Left)",
		keypress.AltArrow{Direction: keypress.Right}: "moveSelection(Right)",

		keypress.AltChar{Char: 'H'}: "moveSelection(Left)",
		keypress.AltChar{Char: 'J'}: "moveSelection(Down)",
		keypress.AltChar{Char: 'K'}: "moveSelection(Up)",
		keypress.AltChar{Char: 'L'}: "moveSelection(Right)",

		keypress.AltShiftChar{'Q'}: "killWindow",

		keypress.AltChar{'R'}: "resize",
	},
}

func seiveConfigEvents(ev interface{}) bool {
	if operationCode, ok := config.bindings[ev]; ok {
		executeOperationCode(operationCode)
		root.simplify()

		root.refreshRenderRect()

		return true
	}

	return false
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

	if root.workspaces[root.selectionIdx].doFullscreen {
		switch funcName {
		case "fullscreen":
			unfullscreen()
		case "killWindow":
			unfullscreen()
			killWindow()
		}
	} else {
		switch funcName {
		case "search":
			search()
		case "fullscreen":
			fullscreen()
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
		case "resize":
			resizeMode = true
		case "debugSlowMode":
			log.Println("slowmo enabled!")
			if getSelection().getContainer().(*Pane).vterm.DebugSlowMode {
				dbug := ""
				scr := getSelection().getContainer().(*Pane).vterm.Screen
				for _, row := range scr {
					for _, ch := range row {
						dbug += string(ch.Rune)
					}
					dbug += "\n"
				}
				log.Println("=== SCREEN OUTPUT ===")
				log.Println(dbug)
				getSelection().getContainer().(*Pane).vterm.DebugSlowMode = false
			} else {
				getSelection().getContainer().(*Pane).vterm.DebugSlowMode = true
			}
		default:
			panic(funcName)
		}

	}
}
