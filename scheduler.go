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

			fmt.Print("\033[?25l")

			escCode := cursor.DeltaMarkup(globalCursor, char.Cursor)
			fmt.Print(escCode)

			if !(char.Cursor.Y > termH && char.Rune == '\n') {
				fmt.Print(string(char.Rune))
			}

			globalCursor = char.Cursor

			if unicode.IsPrint(char.Rune) {
				globalCursor.X++
			}

			desiredCursor := char.Cursor

			t := getSelection().getContainer().(*Term)
			desiredCursor.X = t.vterm.Cursor.X + t.renderRect.x
			desiredCursor.Y = t.vterm.Cursor.Y + t.renderRect.y

			fmt.Print(cursor.DeltaMarkup(globalCursor, desiredCursor))

			fmt.Print("\033[?25h")

			globalCursor = desiredCursor
		case s := <-globalRawAggregate:
			fmt.Print(s)
		}
	}
}
