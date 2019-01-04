package main

// Config stores all user configuration values
type Config struct {
	bindings map[string]func()
}

var config = Config{
	bindings: map[string]func(){
		"Alt+N":           newWindow,
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
