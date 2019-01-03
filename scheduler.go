package main

import (
	"fmt"
	"unicode"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
)

var globalCharAggregate chan vterm.Char

func init() {
	globalCharAggregate = make(chan vterm.Char)
}

func render() {
	lastCursor := cursor.Cursor{}
	for {
		char, ok := <-globalCharAggregate
		if !ok {
			fmt.Println("Exiting scheduler")
			return
		}

		escCode := cursor.DeltaMarkup(lastCursor, char.Cursor)
		fmt.Print(escCode)

		fmt.Print(string(char.Rune))

		lastCursor = char.Cursor

		if unicode.IsPrint(char.Rune) {
			lastCursor.X++
		}
	}
}
