package vterm

import (
	"log"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"
)

// pullRune returns the next byte in the input stream
func (v *VTerm) pullRune() (rune, bool) {
	v.internalByteCounter++

	lag := atomic.LoadUint64(v.shellByteCounter) - v.internalByteCounter
	if lag > uint64(v.w*v.h*4) {
		v.useSlowRefresh()
	} else {
		v.useFastRefresh()
	}

	for {
		select {
		case r, ok := <-v.in:
			if r != 0 {
				if v.DebugSlowMode {
					log.Printf("rune: %v (%s)", r, string(r))
					time.Sleep(100 * time.Millisecond)
				}
			}
			return r, ok
		case p := <-v.ChangePause:
			for {
				v.IsPaused = p
				if !p {
					break
				}
				p = <-v.ChangePause
			}
		}
	}
}

func (v *VTerm) pullRuneNoErr() rune {
	r, _ := v.pullRune()
	return r
}

// ProcessStream processes and transforms a process' stdout, turning it into a stream of Char's to be sent to the rendering scheduler
// This includes translating ANSI Cursor coordinates and maintaining a scrolling buffer
func (v *VTerm) ProcessStream() {
	for {
		next, ok := v.pullRune()
		if !ok {
			return
		}

		if next > 127 {
			value := []byte{byte(next)}

			leadingHex := next >> 4
			switch leadingHex {
			case 12: // 1100
				value = append(value, byte(v.pullRuneNoErr()))
			case 14: // 1110
				value = append(value, byte(v.pullRuneNoErr()))
				value = append(value, byte(v.pullRuneNoErr()))
			case 15: // 1111
				value = append(value, byte(v.pullRuneNoErr()))
				value = append(value, byte(v.pullRuneNoErr()))
				value = append(value, byte(v.pullRuneNoErr()))
			}

			next, _ = utf8.DecodeRune(value)
		}

		switch next {
		case 0:
		case '\x1b':
			v.handleEscapeCode()
		case 8:
			if v.Cursor.X > 0 {
				v.shiftCursorX(-1)
			}
		case '\n':
			if v.Cursor.Y == v.scrollingRegion.bottom {
				v.scrollUp(1)
			} else {
				v.shiftCursorY(1)
			}
		case '\r':
			v.setCursorX(0)
		case '\t':
			tabWidth := 8
			v.Cursor.X += tabWidth - (v.Cursor.X % tabWidth)
		default:
			if unicode.IsPrint(next) {
				if v.Cursor.X < 0 {
					v.setCursorX(0)
				}
				if v.Cursor.Y < 0 {
					v.setCursorY(0)
				}

				if v.Cursor.X >= v.w-1 {
					v.setCursorX(0)
					if v.Cursor.Y == v.scrollingRegion.bottom {
						v.scrollUp(1)
					} else {
						v.shiftCursorY(1)
					}
				}

				v.putChar(next)
			} else {
				log.Printf("Unrecognized unprintable rune: %x", next)
			}
		}
	}
}

func (v *VTerm) handleEscapeCode() {
	next, ok := v.pullRune()
	if !ok {
		return
	}

	switch next {
	case '[':
		v.handleCSISequence()
	case '(': // Character set
		v.pullRune()
		// TODO: implement character sets
	case ']': // operating system command
		// FIXME: are the more terminators than BEL?
		for next != 7 { // slurp until-and-including BEL
			next, ok = v.pullRune()
			if !ok {
				return
			}
		}
	case '=': // Application Keypad
		// TODO
	case 'M': // Reverse Linefeed; http://man7.org/linux/man-pages/man4/console_codes.4.html
		if v.Cursor.Y == 0 {
			v.scrollDown(1)
		} else {
			v.shiftCursorY(-1)
		}
		v.RedrawWindow()
	default:
		log.Printf("Unrecognized escape code: %v", string(next))
	}
}
