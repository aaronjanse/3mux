package pane

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/render"
)

func (t *Pane) doSearch() {
	fullBuffer := append(t.vterm.Scrollback, t.vterm.Screen...)
	match, err := t.locateText(fullBuffer, t.searchText)

	if err == nil {
		t.searchPos = match.y1

		bottomOfScreen := 0
		if match.y1 > t.renderRect.H {
			topOfScreen := match.y1 + t.renderRect.H/2
			if topOfScreen > len(fullBuffer) { // top of scrollback
				topOfScreen = len(fullBuffer) - 1
				t.vterm.ScrollbackPos = len(t.vterm.Scrollback) - 1
			} else {
				t.vterm.ScrollbackPos = topOfScreen - t.renderRect.H - 1
			}
			bottomOfScreen = topOfScreen - t.renderRect.H
			match.y1 -= bottomOfScreen
			match.y2 -= bottomOfScreen
		} else {
			t.vterm.ScrollbackPos = 0
		}

		t.vterm.RedrawWindow()

		for i := match.x1; i <= match.x2; i++ {
			theY := len(fullBuffer) - (bottomOfScreen + match.y1 + 1)
			t.renderer.HandleCh(render.PositionedChar{
				Rune: fullBuffer[theY][i].Rune,
				Cursor: render.Cursor{
					X: t.renderRect.X + i,
					Y: t.renderRect.Y + t.renderRect.H - match.y1,
					Style: render.Style{
						Bg: ecma48.Color{
							ColorMode: ecma48.ColorBit3Bright,
							Code:      2,
						},
						Fg: ecma48.Color{
							ColorMode: ecma48.ColorBit3Normal,
							Code:      0,
						},
					},
				},
			})
		}
	} else {
		log.Println("Could not find match!")
	}
}

// SearchMatch coordinates are relative to bottom left. 1st coords are upper left and 2nd coords are bottom right of search match
type SearchMatch struct {
	x1, y1, x2, y2 int
}

func (t *Pane) locateText(chars [][]render.Char, text string) (SearchMatch, error) {
	lineFromBottom := t.searchPos

	i := len(chars) - t.searchPos - 1
	for {
		var str strings.Builder

		for _, c := range chars[i] {
			str.WriteRune(c.Rune)
		}

		pos := strings.Index(str.String(), text)
		if pos != -1 {
			return SearchMatch{
				x1: pos,
				x2: pos + len(text) - 1,
				y1: lineFromBottom,
				y2: lineFromBottom,
			}, nil
		}
		if t.searchDirection == SearchUp {
			lineFromBottom++
			i--
			if i < 0 {
				break
			}
		} else {
			lineFromBottom--
			i++
			if i >= len(chars) {
				break
			}
		}
	}

	return SearchMatch{}, errors.New("could not find match")
}

func (t *Pane) ToggleSearch() {
	t.searchMode = !t.searchMode

	if t.searchMode {
		t.vterm.ChangePause <- true
		t.searchBackupScrollPos = t.vterm.ScrollbackPos
		t.searchResultsMode = false
		t.searchDirection = SearchUp

		// FIXME hacky way to wait for full control of screen section
		timer := time.NewTimer(time.Millisecond * 5)
		select {
		case <-timer.C:
			timer.Stop()
		}

		lastLineIsBlank := true
		lastLine := t.vterm.Screen[len(t.vterm.Screen)-2]
		for _, c := range lastLine {
			if c.Rune != 32 && c.Rune != 0 {
				lastLineIsBlank = false
				break
			}
		}

		t.searchDidShiftUp = !lastLineIsBlank

		if !lastLineIsBlank {
			blankLine := []render.Char{}
			for i := 0; i < t.renderRect.W; i++ {
				blankLine = append(blankLine, render.Char{Rune: ' ', Style: render.Style{}})
			}

			t.vterm.Scrollback = append(t.vterm.Scrollback, t.vterm.Screen[0])
			t.vterm.Screen = append(t.vterm.Screen[1:], blankLine)

			t.vterm.RedrawWindow()
		}

		t.displayStatusText("Search...")
	} else {
		t.clearStatusText()

		t.vterm.ScrollbackPos = t.searchBackupScrollPos

		if t.searchDidShiftUp {
			t.vterm.Screen = append([][]render.Char{t.vterm.Scrollback[len(t.vterm.Scrollback)-1]}, t.vterm.Screen[:len(t.vterm.Screen)-1]...)
			t.vterm.Scrollback = t.vterm.Scrollback[:len(t.vterm.Scrollback)-1]
		}
		t.vterm.RedrawWindow()
		t.vterm.ChangePause <- false
	}
}

func (t *Pane) displayStatusText(s string) {
	for i := 0; i < t.renderRect.W; i++ {
		r := ' '
		if i < len(s) {
			r = rune(s[i])
		}

		ch := render.PositionedChar{
			Rune: r,
			Cursor: render.Cursor{
				X: t.renderRect.X + i,
				Y: t.renderRect.Y + t.renderRect.H - 1,
				Style: render.Style{
					Bg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Bright,
						Code:      2,
					},
					Fg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		t.renderer.HandleCh(ch)
	}
}

func (t *Pane) clearStatusText() {
	for i := 0; i < t.renderRect.W; i++ {
		ch := render.PositionedChar{
			Rune: ' ',
			Cursor: render.Cursor{
				X: i,
				Y: t.renderRect.H - 1,
				Style: render.Style{
					Bg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Bright,
						Code:      2,
					},
					Fg: ecma48.Color{
						ColorMode: ecma48.ColorBit3Normal,
						Code:      0,
					},
				},
			},
		}
		t.renderer.HandleCh(ch)
	}
}
