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

	"github.com/google/go-cmp/cmp"
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

var framebuffer [][]vterm.Char

func init() {
	var err error
	termW, termH, err = getTermSize()
	if err != nil {
		log.Fatal(err)
	}

	globalCharAggregate = make(chan vterm.Char)
	globalRawAggregate = make(chan string)
	globalCursor = cursor.Cursor{}

	framebuffer = [][]vterm.Char{}
}

func render() {
	for {
		select {
		case char, ok := <-globalCharAggregate:
			if !ok {
				fmt.Println("Exiting scheduler")
				return
			}

			drawChar(char)

			desiredCursor := globalCursor

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

func drawChar(c vterm.Char) {
	if len(framebuffer)-1 < c.Cursor.Y {
		for i := len(framebuffer); i < c.Cursor.Y+1; i++ {
			framebuffer = append(framebuffer, []vterm.Char{vterm.Char{
				Rune: ' ',
			}})
		}
	}
	if len(framebuffer[c.Cursor.Y])-1 < c.Cursor.X {
		for i := len(framebuffer[c.Cursor.Y]); i < c.Cursor.X+1; i++ {
			framebuffer[c.Cursor.Y] = append(framebuffer[c.Cursor.Y], vterm.Char{
				Rune: ' ',
			})
		}
	}

	if !redundant(c) {
		framebuffer[c.Cursor.Y][c.Cursor.X] = c

		escCode := cursor.DeltaMarkup(globalCursor, c.Cursor)
		fmt.Print(escCode)

		fmt.Print(string(c.Rune))

		globalCursor = c.Cursor

		if unicode.IsPrint(c.Rune) {
			globalCursor.X++
		}
	}
}

func redundant(ch vterm.Char) bool {
	c := ch.Cursor
	existing := framebuffer[c.Y][c.X]
	e := existing.Cursor

	if ch.Rune == 0 {
		ch.Rune = ' '
	}

	return ch.Rune == existing.Rune &&
		cmp.Equal(c, e)

	// return ch.Rune == existing.Rune &&
	// 	c.Fg.ColorMode == e.Fg.ColorMode &&
	// 	c.Bg.ColorMode == e.Bg.ColorMode &&
	// 	c.Fg.Code == e.Fg.Code &&
	// 	c.Bg.Code == e.Bg.Code &&
	// 	c.Bold == e.Bold
}
