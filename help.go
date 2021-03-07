package main

import "fmt"

func showHelp() {
	fmt.Println(helpText)
}

const helpText = `3mux
The terminal multiplexer inspired by i3

USAGE:
    3mux                  Interactive 3mux interface
    3mux ls               List session names (has alias '3mux ps')
    3mux attach <name>    Attach to a session
    3mux detach           Detach from the current session
    3mux new <name>       Create a new session
    3mux kill <name>      Kill a session

SHORTCUTS:
	Alt+N/Alt+Enter   Create new pane
	Alt+Shift+Q       Close pane
	Alt+Shift+F       Make pane fullscreen
	Alt+Shift+Arrow   Move pane
	Alt+Arrow         Move selection
	Alt+/             Toggle search
`
