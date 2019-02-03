package main

// Config stores all user configuration values
type Config struct {
	statusBar bool
	bindings  map[string]string
}

var config = Config{
	statusBar: true,
	bindings: map[string]string{
		"Alt+N": "newWindow",

		"Alt+Shift+Up":    "moveWindow(Up)",
		"Alt+Shift+Down":  "moveWindow(Down)",
		"Alt+Shift+Left":  "moveWindow(Left)",
		"Alt+Shift+Right": "moveWindow(Right)",

		"Alt+Shift+I": "moveWindow(Up)",
		"Alt+Shift+K": "moveWindow(Down)",
		"Alt+Shift+J": "moveWindow(Left)",
		"Alt+Shift+L": "moveWindow(Right)",

		"Alt+Up":    "moveSelection(Up)",
		"Alt+Down":  "moveSelection(Down)",
		"Alt+Left":  "moveSelection(Left)",
		"Alt+Right": "moveSelection(Right)",

		"Alt+I": "moveSelection(Up)",
		"Alt+K": "moveSelection(Down)",
		"Alt+J": "moveSelection(Left)",
		"Alt+L": "moveSelection(Right)",

		"Alt+Shift+Q": "killWindow",
	},
}
