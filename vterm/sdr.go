package vterm

import (
	"github.com/aaronduino/i3-tmux/cursor"
)

func (v *VTerm) handleSDR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	if parameterCode == "39;49" {
		v.Cursor.Fg.ColorMode = cursor.ColorNone
		v.Cursor.Bg.ColorMode = cursor.ColorNone
		return
	}

	c := seq[0]

	switch c {
	case 0:
		v.Cursor.Reset()
	case 1:
		v.Cursor.Bold = true
	case 2:
		v.Cursor.Faint = true
	case 3:
		v.Cursor.Italic = true
	case 4:
		v.Cursor.Underline = true
	case 5: // slow blink
	case 6: // rapid blink
	case 7: // swap foreground & background; see case 27
	case 8:
		v.Cursor.Conceal = true
	case 9:
		v.Cursor.CrossedOut = true
	case 10: // primary/default font
	case 22:
		v.Cursor.Bold = false
		v.Cursor.Faint = false
	case 23:
		v.Cursor.Italic = false
	case 24:
		v.Cursor.Underline = false
	case 25: // blink off
	case 27: // inverse off; see case 7
		// TODO
	case 28:
		v.Cursor.Conceal = false
	case 29:
		v.Cursor.CrossedOut = false
	case 38: // set foreground color
		if seq[1] == 5 {
			v.Cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.Cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 39: // default foreground color
		v.Cursor.Fg.ColorMode = cursor.ColorNone
	case 48: // set background color
		if seq[1] == 5 {
			v.Cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.Cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 49: // default background color
		v.Cursor.Bg.ColorMode = cursor.ColorNone
	default:
		if c >= 30 && c <= 37 {
			if len(seq) > 1 && seq[1] == 1 {
				v.Cursor.Fg = cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      int32(c - 30),
				}
			} else {
				v.Cursor.Fg = cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      int32(c - 30),
				}
			}
		} else if c >= 40 && c <= 47 {
			if len(seq) > 1 && seq[1] == 1 {
				v.Cursor.Bg = cursor.Color{
					ColorMode: cursor.ColorBit3Bright,
					Code:      int32(c - 40),
				}
			} else {
				v.Cursor.Bg = cursor.Color{
					ColorMode: cursor.ColorBit3Normal,
					Code:      int32(c - 40),
				}
			}
		} else if c >= 90 && c <= 97 {
			v.Cursor.Fg = cursor.Color{
				ColorMode: cursor.ColorBit3Bright,
				Code:      int32(c - 90),
			}
		} else if c >= 100 && c <= 107 {
			v.Cursor.Bg = cursor.Color{
				ColorMode: cursor.ColorBit3Bright,
				Code:      int32(c - 100),
			}
		} else {
			// v.debug("SGR Code: " + string(parameterCode))
		}
	}
}
