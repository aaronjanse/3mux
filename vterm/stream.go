package vterm

import (
	"log"
	"time"
	"unicode"
	"unicode/utf8"
)

// ProcessStream processes and transforms a process' stdout, turning it into a stream of Char's to be sent to the rendering scheduler
// This includes translating ANSI Cursor coordinates and maintaining a scrolling buffer
func (v *VTerm) ProcessStream() {
	// f, err := os.OpenFile("vterm-logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)

	for {
		next, ok := <-v.in
		if !ok {
			return
		}

		if next == '@' {
			nowTime := time.Now().UnixNano()
			log.Printf("%v ms - time to stream finish\n", (nowTime-v.startTime)/1000000)
			v.shutdown <- true
			return
		}

		// <-time.NewTimer(time.Second / 32).C

		if next > 127 {
			value := []byte{byte(next)}

			leadingHex := next >> 4
			switch leadingHex {
			case 12: // 1100
				value = append(value, byte(<-v.in))
			case 14: // 1110
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
			case 15: // 1111
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
				value = append(value, byte(<-v.in))
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
	next, ok := <-v.in
	if !ok {
		log.Fatal("not ok")
		return
	}

	switch next {
	case '[':
		v.handleCSISequence()
	case '(': // Character set
		<-v.in
		// TODO
	default:
		// v.debug("ESC Code: " + string(next))
	}
}
