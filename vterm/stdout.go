package vterm

import (
	"bufio"
	"log"
	"sync/atomic"

	"github.com/aaronjanse/3mux/ecma48"
)

type Parser struct {
	state State

	private      *rune
	intermediate string
	params       string
	final        *rune
}

type State int

const (
	StateGround = iota
	StateEscape
	StateCsiEntry
	StateCsiParam
	StateOscString
)

func (v *VTerm) ProcessStdout(input *bufio.Reader) {
	stdout := make(chan ecma48.Output, 3200000)
	shutdown := make(chan bool)

	parser := ecma48.NewParser(false)

	go func() {
		parser.Parse(input, stdout)
		shutdown <- true
	}()

	for {
		select {
		case p := <-v.ChangePause:
			for {
				v.IsPaused = p
				if !p {
					break
				}
				p = <-v.ChangePause
			}
		case <-shutdown:
			return
		case output := <-stdout:
			v.runeCounter += uint64(len(output.Raw))

			lag := atomic.LoadUint64(&parser.RuneCounter) - v.runeCounter
			if lag > uint64(v.w*v.h) {
				v.useSlowRefresh()
			} else {
				v.useFastRefresh()
			}

			// log.Printf(":: %q", output.Raw)

			switch x := output.Parsed.(type) {
			case ecma48.Char:
				v.putChar(x.Rune, x.IsWide)
			case ecma48.Backspace:
				if v.Cursor.X > 0 {
					v.shiftCursorX(-1)
				}
				v.RedrawWindow()
			case ecma48.Newline:
				if v.Cursor.Y == v.scrollingRegion.bottom-1 {
					v.scrollUp(1)
				} else if v.Cursor.Y < v.h {
					v.shiftCursorY(1)
				}
			case ecma48.RI:
				if v.Cursor.Y == v.scrollingRegion.top {
					v.scrollDown(1)
				} else if v.Cursor.Y > 0 {
					v.shiftCursorY(-1)
				}
			case ecma48.CarriageReturn:
				v.setCursorX(0)
			case ecma48.Tab:
				tabWidth := 8 // FIXME
				v.shiftCursorX(tabWidth - (v.Cursor.X % tabWidth))
			case ecma48.ICH: // insert characters
				if v.Cursor.X+x.N >= v.w {
					x.N = v.w - v.Cursor.X - 1
				}
				copy(v.Screen[v.Cursor.Y][v.Cursor.X+x.N:], v.Screen[v.Cursor.Y][v.Cursor.X:])
				for i := 0; i < x.N; i++ {
					v.Screen[v.Cursor.Y][v.Cursor.X+i] = ecma48.StyledChar{
						Rune: ' ', IsWide: false, Style: v.Cursor.Style,
					}
				}

				v.RedrawWindow() // FIXME inefficient
			case ecma48.DCH: // delete characters
				if x.N > v.w-v.Cursor.X {
					x.N = v.w - v.Cursor.X // FIXME: verify that we don't need +/- 1
				}
				copy(v.Screen[v.Cursor.Y][v.Cursor.X:], v.Screen[v.Cursor.Y][v.Cursor.X+x.N:])
				for i := 0; i < x.N; i++ {
					if v.w-1-i >= len(v.Screen[v.Cursor.Y]) {
						continue
					}
					v.Screen[v.Cursor.Y][v.w-1-i] = ecma48.StyledChar{
						Rune: ' ', IsWide: false, Style: v.Cursor.Style,
					}
				}

				v.RedrawWindow() // FIXME inefficient
			case ecma48.PrivateDEC:
				switch x.Code {
				// FIXME: distinguish between these
				case 1049, 1047, 47:
					if x.On {
						if !v.UsingAltScreen {
							// TODO: reshape if needed
							v.screenBackup = v.Screen
						}
					} else {
						if v.UsingAltScreen {
							v.Screen = v.screenBackup
						}
					}
					v.UsingAltScreen = x.On
				default:
					log.Printf("Unrecognized DEC Private Mode: %d", x.Code)
				}

			case ecma48.CursorMovement:
				switch x.Direction {
				case ecma48.Up:
					v.shiftCursorY(-x.N)
				case ecma48.Down:
					v.shiftCursorY(x.N)
				case ecma48.Left:
					v.shiftCursorX(-x.N)
				case ecma48.Right:
					v.shiftCursorX(x.N)
				}

			case ecma48.VPA:
				v.setCursorY(x.Y)
			case ecma48.CNL:
				v.shiftCursorY(int(x.YDiff))
				v.setCursorX(0)
			case ecma48.CPL:
				v.shiftCursorY(-int(x.YDiff))
				v.setCursorX(0)
			case ecma48.CHA:
				v.shiftCursorX(x.X)
			case ecma48.CUP:
				v.setCursorPos(x.X, x.Y)
			case ecma48.ED:
				v.handleEraseInDisplay(x.Directive)
			case ecma48.EL:
				v.handleEraseInLine(x.Directive)
			case ecma48.IL:
				if v.Cursor.Y < v.scrollingRegion.top {
					break
				}
				if v.Cursor.Y >= v.scrollingRegion.bottom {
					break
				}

				if x.N > v.scrollingRegion.bottom-v.Cursor.Y {
					x.N = v.scrollingRegion.bottom - v.Cursor.Y
				}

				newLines := make([][]ecma48.StyledChar, x.N)
				for i := range newLines {
					newLines[i] = make([]ecma48.StyledChar, v.w)
					// if we're at the top, use *blank* lines
					if v.Cursor.Y == v.scrollingRegion.top {
						for x := range newLines[i] {
							newLines[i][x].Style = v.Cursor.Style
						}
					}
				}

				newLines = append(append(
					newLines,
					v.Screen[v.Cursor.Y:v.scrollingRegion.bottom-x.N]...),
					v.Screen[v.scrollingRegion.bottom:]...)

				copy(v.Screen[v.Cursor.Y:], newLines)

				v.RedrawWindow()
			case ecma48.DL:
				if v.Cursor.Y < v.scrollingRegion.top {
					break
				}
				if v.Cursor.Y >= v.scrollingRegion.bottom {
					break
				}

				if x.N > v.scrollingRegion.bottom-v.Cursor.Y {
					x.N = v.scrollingRegion.bottom - v.Cursor.Y
				}

				newLines := make([][]ecma48.StyledChar, x.N)
				for i := range newLines {
					newLines[i] = make([]ecma48.StyledChar, v.w)
					for x := range newLines[i] {
						newLines[i][x].Style = v.Cursor.Style
					}
				}

				v.Screen = append(append(append(
					v.Screen[:v.Cursor.Y],
					v.Screen[v.Cursor.Y+x.N:v.scrollingRegion.bottom]...),
					newLines...),
					v.Screen[v.scrollingRegion.bottom:]...)

				if !v.usingSlowRefresh {
					v.RedrawWindow()
				}
			case ecma48.DECSTBM:
				if x.Top < 1 {
					v.scrollingRegion.top = 0
				} else if x.Top > v.h {
					v.scrollingRegion.top = v.h - 1
				} else {
					v.scrollingRegion.top = x.Top - 1
				}
				if x.Bottom == -1 || x.Bottom >= v.h {
					v.scrollingRegion.bottom = v.h
				} else {
					v.scrollingRegion.bottom = x.Bottom + 1
				}
				if v.scrollingRegion.bottom <= v.scrollingRegion.top {
					v.scrollingRegion.bottom = v.scrollingRegion.top + 1
				}
				v.setCursorPos(0, 0)
			case ecma48.SU:
				v.scrollUp(int(x.N))
			case ecma48.SD:
				v.scrollDown(int(x.N))
			case ecma48.SCOSC:
				v.storedCursorX = v.Cursor.X
				v.storedCursorY = v.Cursor.Y
			case ecma48.SCORC:
				v.setCursorPos(v.storedCursorX, v.storedCursorY)

			case ecma48.StyleReset:
				v.Cursor.Style.Reset()

			case ecma48.StyleForeground:
				v.Cursor.Style.Fg = ecma48.Color(x)
			case ecma48.StyleBackground:
				v.Cursor.Style.Bg = ecma48.Color(x)

			case ecma48.StyleBold:
				v.Cursor.Style.Bold = bool(x)
			case ecma48.StyleConceal:
				v.Cursor.Style.Conceal = bool(x)
			case ecma48.StyleCrossedOut:
				v.Cursor.Style.CrossedOut = bool(x)
			case ecma48.StyleItalic:
				v.Cursor.Style.Italic = bool(x)
			case ecma48.StyleFaint:
				v.Cursor.Style.Faint = bool(x)
			case ecma48.StyleReverse:
				v.Cursor.Style.Reverse = bool(x)
			case ecma48.StyleUnderline:
				v.Cursor.Style.Underline = bool(x)

			case ecma48.Unrecognized:
				log.Printf("?? %q", output.Raw)
			default:
				log.Printf("Unrecognized parser output: %+v", x)
			}
		}
	}
}
