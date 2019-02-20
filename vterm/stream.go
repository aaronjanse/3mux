package vterm

import (
	"log"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
)

// pullByte returns the next byte in the input stream
func (v *VTerm) pullByte() (rune, bool) {
	v.internalByteCounter++

	lag := atomic.LoadUint64(v.shellByteCounter) - v.internalByteCounter
	if lag > uint64(v.w*v.h*4) {
		v.useSlowRefresh()
	} else {
		v.useFastRefresh()
	}

	r, ok := <-v.in
	return r, ok
}

func (v *VTerm) pullByteNoErr() rune {
	r, _ := v.pullByte()
	return r
}

// ProcessStream processes and transforms a process' stdout, turning it into a stream of Char's to be sent to the rendering scheduler
// This includes translating ANSI Cursor coordinates and maintaining a scrolling buffer
func (v *VTerm) ProcessStream() {
	for {
		next, ok := v.pullByte()
		if !ok {
			return
		}

		// if next == '@' {
		// 	nowTime := time.Now().UnixNano()
		// 	log.Printf("%v ms - time to stream finish\n", (nowTime-v.startTime)/1000000)
		// 	v.RedrawWindow()
		// 	// time.Sleep(time.Second)
		// 	// v.shutdown <- true
		// 	// return
		// }

		if next > 127 {
			value := []byte{byte(next)}

			leadingHex := next >> 4
			switch leadingHex {
			case 12: // 1100
				value = append(value, byte(v.pullByteNoErr()))
			case 14: // 1110
				value = append(value, byte(v.pullByteNoErr()))
				value = append(value, byte(v.pullByteNoErr()))
			case 15: // 1111
				value = append(value, byte(v.pullByteNoErr()))
				value = append(value, byte(v.pullByteNoErr()))
				value = append(value, byte(v.pullByteNoErr()))
			}

			next, _ = utf8.DecodeRune(value)
		}

		switch next {
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
			tabWidth := 4
			v.Cursor.X += tabWidth - (v.Cursor.X % tabWidth)
		default:
			if unicode.IsPrint(next) {
				if v.Cursor.X < 0 {
					v.setCursorX(0)
				}
				if v.Cursor.Y < 0 {
					v.setCursorY(0)
				}

				v.putChar(next)
			} else {
				// v.debug(fmt.Sprintf("%x    ", next))
			}
		}
	}
}

func (v *VTerm) handleEscapeCode() {
	next, ok := v.pullByte()
	if !ok {
		log.Fatal("not ok")
		return
	}

	switch next {
	case '[':
		v.handleCSISequence()
	case '(': // Character set
		v.pullByte()
		// TODO
	default:
		// v.debug("ESC Code: " + string(next))
	}
}
