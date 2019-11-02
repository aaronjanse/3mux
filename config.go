package main

import (
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
		keypress.AltChar{'N'}: "newWindow",

		keypress.AltShiftArrow{keypress.Up}:    "moveWindow(Up)",
		keypress.AltShiftArrow{keypress.Down}:  "moveWindow(Down)",
		keypress.AltShiftArrow{keypress.Left}:  "moveWindow(Left)",
		keypress.AltShiftArrow{keypress.Right}: "moveWindow(Right)",

		keypress.AltShiftChar{'I'}: "moveWindow(Up)",
		keypress.AltShiftChar{'K'}: "moveWindow(Down)",
		keypress.AltShiftChar{'J'}: "moveWindow(Left)",
		keypress.AltShiftChar{'L'}: "moveWindow(Right)",

		keypress.AltArrow{keypress.Up}:    "moveSelection(Up)",
		keypress.AltArrow{keypress.Down}:  "moveSelection(Down)",
		keypress.AltArrow{keypress.Left}:  "moveSelection(Left)",
		keypress.AltArrow{keypress.Right}: "moveSelection(Right)",

		keypress.AltChar{'I'}: "moveSelection(Up)",
		keypress.AltChar{'K'}: "moveSelection(Down)",
		keypress.AltChar{'J'}: "moveSelection(Left)",
		keypress.AltChar{'L'}: "moveSelection(Right)",

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
	case "resize":
		resizeMode = true
	default:
		panic(funcName)
	}
}
