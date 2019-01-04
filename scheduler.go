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
	// // blink cursor
	// cursorTicker := time.NewTicker(time.Second / 2)
	// cursorDone := make(chan bool)
	// defer (func() {
	// 	cursorDone <- true
	// })()
	// go (func() {
	// 	visible := true
	// 	for {
	// 		select {
	// 		case <-cursorTicker.C:
	// 			t := getSelection().getContainer().(*Term)

	// 			v.bufferMutux.Lock()
	// 			char := v.buffer[v.Cursor.Y][v.Cursor.X]
	// 			if visible && v.Selected {
	// 				char.Cursor.Underline = true
	// 			}
	// 			v.out <- char

	// 			v.bufferMutux.Unlock()
	// 			visible = !visible
	// 			break
	// 		case <-cursorDone:
	// 			cursorTicker.Stop()
	// 			return
	// 		}
	// 	}
	// })()

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
