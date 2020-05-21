package pane

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/mattn/go-runewidth"
)

// SearchDirection is which direction we move through search results
type SearchDirection int

// enum of search directions
const (
	SearchUp SearchDirection = iota
	SearchDown
)

func (t *Pane) handleSearchStdin(in string) {
	if t.searchResultsMode {
		switch in[0] { // FIXME ignores extra chars
		case 'n': // next
			t.searchDirection = SearchDown
			t.searchPos--
			if t.searchPos < 0 {
				t.searchPos = 0
			}
			t.doSearch()
		case 'N': // prev
			t.searchDirection = SearchUp
			t.searchPos++
			max := len(t.vterm.Scrollback) + len(t.vterm.Screen) - 1
			if t.searchPos > max {
				t.searchPos = max
			}
			t.doSearch()
		case '/':
			t.searchResultsMode = false
			t.displayStatusText(t.searchText)
		case 127:
			fallthrough
		case 8:
			t.searchResultsMode = false
			t.searchText = t.searchText[:len(t.searchText)-1]
			t.displayStatusText(t.searchText)
		case 3:
			fallthrough
		case 4:
			fallthrough
		case 13:
			fallthrough
		case 10: // enter
			t.ToggleSearch()
			t.vterm.ScrollbackPos = t.searchPos - len(t.vterm.Screen) + t.renderRect.H/2
			t.vterm.RedrawWindow()
		}
	} else {
		for _, c := range in {
			if c == 3 || c == 4 || c == 27 {
				t.ToggleSearch()
				return
			} else if c == 8 || c == 127 { // backspace
				if len(t.searchText) > 0 {
					t.searchText = t.searchText[:len(t.searchText)-1]
				}
			} else if c == 10 || c == 13 {
				if len(t.searchText) == 0 {
					t.ToggleSearch()
					return
				} else {
					t.searchResultsMode = true
					return // FIXME ignores extra chars
				}
			} else {
				t.searchText += string(c)
			}
		}
		t.searchPos = 0
		t.doSearch()
		t.displayStatusText(t.searchText)
	}
}

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
			t.renderer.HandleCh(ecma48.PositionedChar{
				Rune: fullBuffer[theY][i].Rune,
				Cursor: ecma48.Cursor{
					X: t.renderRect.X + i,
					Y: t.renderRect.Y + t.renderRect.H - match.y1,
					Style: ecma48.Style{
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

func (t *Pane) locateText(chars [][]ecma48.StyledChar, text string) (SearchMatch, error) {
	lineFromBottom := t.searchPos

	i := len(chars) - t.searchPos - 1
	for {
		var strB strings.Builder

		for _, c := range chars[i] {
			r := c.Rune
			if r == '\x00' {
				r = ' '
			}
			strB.WriteRune(r)
		}

		str := strB.String()
		pos := strings.Index(str, text)
		if pos != -1 {
			pos = runewidth.StringWidth(str[:pos])
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
			blankLine := []ecma48.StyledChar{}
			for i := 0; i < t.renderRect.W; i++ {
				blankLine = append(blankLine, ecma48.StyledChar{Rune: ' ', Style: ecma48.Style{}})
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
			t.vterm.Screen = append([][]ecma48.StyledChar{t.vterm.Scrollback[len(t.vterm.Scrollback)-1]}, t.vterm.Screen[:len(t.vterm.Screen)-1]...)
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

		ch := ecma48.PositionedChar{
			Rune: r,
			Cursor: ecma48.Cursor{
				X: t.renderRect.X + i,
				Y: t.renderRect.Y + t.renderRect.H - 1,
				Style: ecma48.Style{
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
		ch := ecma48.PositionedChar{
			Rune: ' ',
			Cursor: ecma48.Cursor{
				X: i,
				Y: t.renderRect.H - 1,
				Style: ecma48.Style{
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
