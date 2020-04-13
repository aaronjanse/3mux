package vterm

import (
	"log"
)

func (v *VTerm) handleSGR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	if parameterCode == "39;49" {
		v.Cursor.Style.Fg.ColorMode = ColorNone
		v.Cursor.Style.Bg.ColorMode = ColorNone
		return
	}

	for {
		c := seq[0]

		switch c {
		case 0:
			v.Cursor.Style = Style{}
			seq = seq[1:]
		case 1:
			v.Cursor.Style.Bold = true
			seq = seq[1:]
		case 2:
			v.Cursor.Style.Faint = true
			seq = seq[1:]
		case 3:
			v.Cursor.Style.Italic = true
			seq = seq[1:]
		case 4:
			v.Cursor.Style.Underline = true
			seq = seq[1:]
		case 5: // slow blink
			seq = seq[1:]
		case 6: // rapid blink
			seq = seq[1:]
		case 7: // swap foreground & background; see case 27
			seq = seq[1:]
		case 8:
			v.Cursor.Style.Conceal = true
			seq = seq[1:]
		case 9:
			v.Cursor.Style.CrossedOut = true
			seq = seq[1:]
		case 10: // primary/default font
			v.Cursor.Style.Underline = false
			seq = seq[1:]
		case 22:
			v.Cursor.Style.Bold = false
			v.Cursor.Style.Faint = false
			seq = seq[1:]
		case 23:
			v.Cursor.Style.Italic = false
			seq = seq[1:]
		case 24:
			v.Cursor.Style.Underline = false
			seq = seq[1:]
		case 25: // blink off
			seq = seq[1:]
		case 27: // inverse off; see case 7
			seq = seq[1:]
		case 28:
			v.Cursor.Style.Conceal = false
			seq = seq[1:]
		case 29:
			v.Cursor.Style.CrossedOut = false
			seq = seq[1:]

		case 38: // set foreground color
			if seq[1] == 5 {
				v.Cursor.Style.Fg = Color{
					ColorMode: ColorBit8,
					Code:      int32(seq[2]),
				}
				seq = seq[3:]
			} else if seq[1] == 2 {
				v.Cursor.Style.Fg = Color{
					ColorMode: ColorBit24,
					Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
				}
				seq = seq[5:]
			}
		case 39: // default foreground color
			v.Cursor.Style.Fg.ColorMode = ColorNone
			seq = seq[1:]
		case 48: // set background color
			if seq[1] == 5 {
				v.Cursor.Style.Bg = Color{
					ColorMode: ColorBit8,
					Code:      int32(seq[2]),
				}
				seq = seq[3:]
			} else if seq[1] == 2 {
				v.Cursor.Style.Bg = Color{
					ColorMode: ColorBit24,
					Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
				}
				seq = seq[5:]
			}
		case 49: // default background color
			v.Cursor.Style.Bg.ColorMode = ColorNone
			seq = seq[1:]
		default:
			var colorMode ColorMode
			var code int32
			var bg bool

			if c >= 30 && c <= 37 {
				bg = false
				code = int32(c - 30)
				if len(seq) > 1 && seq[1] == 1 {
					colorMode = ColorBit3Bright
				} else {
					colorMode = ColorBit3Normal
				}
				seq = seq[1:]
			} else if c >= 40 && c <= 47 {
				bg = true
				code = int32(c - 40)
				if len(seq) > 1 && seq[1] == 1 {
					colorMode = ColorBit3Bright
				} else {
					colorMode = ColorBit3Normal
				}
				seq = seq[1:]
			} else if c >= 90 && c <= 97 {
				bg = false
				code = int32(c - 90)
				colorMode = ColorBit3Bright
				seq = seq[1:]
			} else if c >= 100 && c <= 107 {
				bg = true
				code = int32(c - 100)
				colorMode = ColorBit3Bright
				seq = seq[1:]
			} else {
				log.Printf("Unrecognized SGR code: %v", parameterCode)
			}

			color := Color{ColorMode: colorMode, Code: code}
			if bg {
				v.Cursor.Style.Bg = color
			} else {
				v.Cursor.Style.Fg = color
			}
		}

		if len(seq) == 0 {
			break
		}
	}
}
