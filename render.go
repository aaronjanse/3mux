package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	"github.com/aaronduino/i3-tmux/cursor"
	"github.com/aaronduino/i3-tmux/vterm"
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

// getTermSize returns the wusth
func getTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}

var termW, termH int

var globalCharAggregate chan vterm.Char
var globalRawAggregate chan string

var globalCursor cursor.Cursor

func init() {
	var err error
	termW, termH, err = getTermSize()
	if err != nil {
		log.Fatal(err)
	}

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

			fmt.Print("\033[?25l") // hide cursor

			/* draw the character we just got */

			escCode := cursor.DeltaMarkup(globalCursor, char.Cursor)
			fmt.Print(escCode)

			if !(char.Cursor.Y > termH && char.Rune == '\n') {
				fmt.Print(string(char.Rune))
			}

			globalCursor = char.Cursor

			if unicode.IsPrint(char.Rune) {
				globalCursor.X++
			}

			/* move the cursor to the correct place in the selected pane */

			desiredCursor := char.Cursor

			t := getSelection().getContainer().(*Term)
			desiredCursor.X = t.vterm.Cursor.X + t.renderRect.x
			desiredCursor.Y = t.vterm.Cursor.Y + t.renderRect.y

			fmt.Print(cursor.DeltaMarkup(globalCursor, desiredCursor))

			fmt.Print("\033[?25h") // show cursor

			globalCursor = desiredCursor
		case s := <-globalRawAggregate:
			fmt.Print(s)
		}
	}
}

func refreshEverything() {
	// fmt.Print("\033[2J")

	var h int
	if config.statusBar {
		h = termH - 1
	} else {
		h = termH
	}
	root.setRenderRect(0, 0, termW, h)

	if config.statusBar {
		debug(root.serialize())
	}
}
