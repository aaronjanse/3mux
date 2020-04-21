package main

import (
	"fmt"
	"strings"

	"github.com/aaronjanse/3mux/wm"
)

// Config stores all user configuration values
type Config struct {
	statusBar bool
	bindings  map[string]func(*wm.Universe)
	modes     []string
}

var configFuncBindings = map[string]func(*wm.Universe){
	"newPane":      func(u *wm.Universe) { u.AddPane() },
	"newPaneHoriz": func(u *wm.Universe) { u.AddPaneTmux(false) },
	"newPaneVert":  func(u *wm.Universe) { u.AddPaneTmux(true) },

	"killWindow": func(u *wm.Universe) { u.KillPane() },
	"fullscreen": func(u *wm.Universe) { u.ToggleFullscreen() },
	"search":     func(u *wm.Universe) { u.ToggleSearch() },

	"resize(Up)":    func(u *wm.Universe) { u.ResizePane(wm.Up) },
	"resize(Down)":  func(u *wm.Universe) { u.ResizePane(wm.Down) },
	"resize(Left)":  func(u *wm.Universe) { u.ResizePane(wm.Left) },
	"resize(Right)": func(u *wm.Universe) { u.ResizePane(wm.Right) },

	"moveWindow(Up)":    func(u *wm.Universe) { u.MoveWindow(wm.Up) },
	"moveWindow(Down)":  func(u *wm.Universe) { u.MoveWindow(wm.Down) },
	"moveWindow(Left)":  func(u *wm.Universe) { u.MoveWindow(wm.Left) },
	"moveWindow(Right)": func(u *wm.Universe) { u.MoveWindow(wm.Right) },

	"moveSelection(Up)":    func(u *wm.Universe) { u.MoveSelection(wm.Up) },
	"moveSelection(Down)":  func(u *wm.Universe) { u.MoveSelection(wm.Down) },
	"moveSelection(Left)":  func(u *wm.Universe) { u.MoveSelection(wm.Left) },
	"moveSelection(Right)": func(u *wm.Universe) { u.MoveSelection(wm.Right) },

	"cycleSelection(Forward)":  func(u *wm.Universe) { u.CycleSelection(true) },
	"cycleSelection(Backward)": func(u *wm.Universe) { u.CycleSelection(false) },

	"splitPane(Vertical)":   func(u *wm.Universe) { u.AddPaneTmux(true) },
	"splitPane(Horizontal)": func(u *wm.Universe) { u.AddPaneTmux(false) },
}

func compileBindings(sourceBindings map[string][]string) ([]string, map[string]func(*wm.Universe)) {
	modes := []string{}
	compiledBindings := map[string]func(*wm.Universe){}
	for funcName, keyCodes := range sourceBindings {
		fn, ok := configFuncBindings[funcName]
		if !ok {
			panic("Incorrect keybinding: " + funcName)
		}
		for _, keyCode := range keyCodes {
			compiledBindings[strings.ToLower(keyCode)] = fn

			parts := strings.Split(keyCode, " ")
			switch len(parts) {
			case 1:
				// ignore
			case 2:
				modes = append(modes, strings.ToLower(parts[0]))
			default:
				panic(fmt.Sprintf("Unexpected config shortcut part count in: %s", keyCode))
			}
		}
	}

	return modes, compiledBindings
}

var config = Config{
	statusBar: true,
}

func init() {
	modes, bindings := compileBindings(map[string][]string{
		"newPane":      []string{`Alt+N`, `Alt+Enter`},
		"newPaneHoriz": []string{`Ctrl+B "`},
		"newPaneVert":  []string{`Ctrl+B %`},

		"killWindow": []string{`Alt+Shift+Q`},
		"fullscreen": []string{`Alt+Shift+F`},
		"search":     []string{`Alt+/`},

		"resize(Up)":    []string{`Alt+R Up`},
		"resize(Down)":  []string{`Alt+R Down`},
		"resize(Left)":  []string{`Alt+R Left`},
		"resize(Right)": []string{`Alt+R Right`},

		"moveWindow(Up)":    []string{`Alt+Shift+K`, `Alt+Shift+Up`},
		"moveWindow(Down)":  []string{`Alt+Shift+J`, `Alt+Shift+Down`},
		"moveWindow(Left)":  []string{`Alt+Shift+H`, `Alt+Shift+Left`, `Ctrl+B {`},
		"moveWindow(Right)": []string{`Alt+Shift+L`, `Alt+Shift+Right`, `Ctrl+B }`},

		"moveSelection(Up)":    []string{`Alt+K`, `Alt+Up`},
		"moveSelection(Down)":  []string{`Alt+J`, `Alt+Down`},
		"moveSelection(Left)":  []string{`Alt+H`, `Alt+Left`},
		"moveSelection(Right)": []string{`Alt+L`, `Alt+Right`},
	})

	config.modes = modes
	config.bindings = bindings
}

var mode = ""

func seiveConfigEvents(u *wm.Universe, human string) bool {
	hu := strings.ToLower(human)
	if mode == "" {
		for _, possibleMode := range config.modes {
			if hu == possibleMode {
				mode = hu
				return true
			}
		}

		if fn, ok := config.bindings[hu]; ok {
			fn(u)
			return true
		}
	} else {
		code := mode + " " + hu
		mode = ""

		if fn, ok := config.bindings[code]; ok {
			fn(u)
			return true
		}

	}
	return false
}
