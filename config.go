package main

import (
	"log"
	"strings"
)

// Config stores all user configuration values
type Config struct {
	statusBar bool
	bindings  map[string]func()
}

var configFuncBindings = map[string]func(){
	"newWindow": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			root.AddPane()
			root.simplify()
			root.refreshRenderRect()
		}
	},
	"killWindow": func() {
		if root.workspaces[root.selectionIdx].doFullscreen {
			unfullscreen()
		}
		killWindow()
		root.simplify()
		root.refreshRenderRect()
	},
	"resize": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			resizeMode = true
		}
	},
	"fullscreen": func() {
		if root.workspaces[root.selectionIdx].doFullscreen {
			unfullscreen()
		} else {
			fullscreen()
		}
		root.simplify()
		root.refreshRenderRect()
	},
	"debugSlowMode": func() {
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
	},
	"search": search,
	"moveWindowUp": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveWindow(Up)
			root.simplify()
			root.refreshRenderRect()
		}
	},
	"moveWindowDown": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveWindow(Down)
			root.simplify()
			root.refreshRenderRect()
		}
	},
	"moveWindowLeft": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveWindow(Left)
			root.simplify()
			root.refreshRenderRect()
		}
	},
	"moveWindowRight": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveWindow(Right)
		}
	},
	"moveSelectionUp": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveSelection(Up)
		}
	},
	"moveSelectionDown": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveSelection(Down)
		}
	},
	"moveSelectionLeft": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveSelection(Left)
		}
	},
	"moveSelectionRight": func() {
		if !root.workspaces[root.selectionIdx].doFullscreen {
			moveSelection(Right)
		}
	},
}

func compileBindings(sourceBindings map[string][]string) map[string]func() {
	compiledBindings := map[string]func(){}
	for funcName, keyCodes := range sourceBindings {
		fn := configFuncBindings[funcName]
		for _, keyCode := range keyCodes {
			compiledBindings[keyCode] = fn
		}
	}

	return compiledBindings
}

var config = Config{
	statusBar: true,
}

func init() {
	config.bindings = compileBindings(map[string][]string{
		"newWindow":     []string{"Alt+N", "Alt+Enter"},
		"killWindow":    []string{"Alt+Shift+Q"},
		"resize":        []string{"Alt+R"},
		"fullscreen":    []string{"Alt+Shift+F"},
		"debugSlowMode": []string{"Alt+X"},
		"search":        []string{"Alt+/"},

		"moveWindowUp":    []string{"Alt+Shift+K", "Alt+Shift+Up"},
		"moveWindowDown":  []string{"Alt+Shift+J", "Alt+Shift+Down"},
		"moveWindowLeft":  []string{"Alt+Shift+H", "Alt+Shift+Left"},
		"moveWindowRight": []string{"Alt+Shift+L", "Alt+Shift+Right"},

		"moveSelectionUp":    []string{"Alt+K", "Alt+Up"},
		"moveSelectionDown":  []string{"Alt+J", "Alt+Down"},
		"moveSelectionLeft":  []string{"Alt+H", "Alt+Left"},
		"moveSelectionRight": []string{"Alt+L", "Alt+Right"},
	})
}

func seiveConfigEvents(human string) bool {
	if fn, ok := config.bindings[human]; ok {
		fn()
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
		case "search":
			search()
		}
	} else {
		switch funcName {
		case "search":
			search()
		case "fullscreen":
			fullscreen()
		case "newWindow":
			root.AddPane()
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
