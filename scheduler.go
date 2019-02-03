package main

import (
	"fmt"
	"unicode"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
)

var globalCharAggregate chan vterm.Char
var globalRawAggregate chan string

var globalCursor cursor.Cursor

func init() {
	globalCharAggregate = make(chan vterm.Char)
	globalRawAggregate = make(chan string)
	globalCursor = cursor.Cursor{}
}

func render() {
	for {
		select {
		case char, ok := <-globalCharAggregate:
			if !ok {
				fmt.Println("Exiting scheduler")
				return
			}

			escCode := cursor.DeltaMarkup(globalCursor, char.Cursor)
			fmt.Print(escCode)

			fmt.Print(string(char.Rune))

			globalCursor = char.Cursor

			if unicode.IsPrint(char.Rune) {
				globalCursor.X++
			}
		case s := <-globalRawAggregate:
			fmt.Print(s)
		}
	}
}
