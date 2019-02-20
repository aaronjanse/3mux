package vterm

import (
	"log"

	"github.com/aaronduino/i3-tmux/render"
)

func (v *VTerm) handleSGR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	if parameterCode == "39;49" {
		v.Cursor.Style.Fg.ColorMode = render.ColorNone
		v.Cursor.Style.Bg.ColorMode = render.ColorNone
		return
	}

	c := seq[0]

	switch c {
	case 0:
		v.Cursor.Style.Reset()
	case 1:
		v.Cursor.Style.Bold = true
	case 2:
		v.Cursor.Style.Faint = true
	case 3:
		v.Cursor.Style.Italic = true
	case 4:
		v.Cursor.Style.Underline = true
	case 5: // slow blink
	case 6: // rapid blink
	case 7: // swap foreground & background; see case 27
	case 8:
		v.Cursor.Style.Conceal = true
	case 9:
		v.Cursor.Style.CrossedOut = true
	case 10: // primary/default font
		v.Cursor.Style.Underline = false
	case 22:
		v.Cursor.Style.Bold = false
		v.Cursor.Style.Faint = false
	case 23:
		v.Cursor.Style.Italic = false
	case 24:
		v.Cursor.Style.Underline = false
	case 25: // blink off
	case 27: // inverse off; see case 7
	case 28:
		v.Cursor.Style.Conceal = false
	case 29:
		v.Cursor.Style.CrossedOut = false
	case 38: // set foreground color
		if seq[1] == 5 {
			v.Cursor.Style.Fg = render.Color{
				ColorMode: render.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.Cursor.Style.Fg = render.Color{
				ColorMode: render.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 39: // default foreground color
		v.Cursor.Style.Fg.ColorMode = render.ColorNone
	case 48: // set background color
		if seq[1] == 5 {
			v.Cursor.Style.Bg = render.Color{
				ColorMode: render.ColorBit8,
				Code:      int32(seq[2]),
			}
		} else if seq[1] == 2 {
			v.Cursor.Style.Bg = render.Color{
				ColorMode: render.ColorBit24,
				Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
			}
		}
	case 49: // default background color
		v.Cursor.Style.Bg.ColorMode = render.ColorNone
	default:
		var colorMode render.ColorMode
		var code int32
		var bg bool

		if c >= 30 && c <= 37 {
			bg = false
			code = int32(c - 30)
			if len(seq) > 1 && seq[1] == 1 {
				colorMode = render.ColorBit3Bright
			} else {
				colorMode = render.ColorBit3Normal
			}
		} else if c >= 40 && c <= 47 {
			bg = true
			code = int32(c - 40)
			if len(seq) > 1 && seq[1] == 1 {
				colorMode = render.ColorBit3Bright
			} else {
				colorMode = render.ColorBit3Normal
			}
		} else if c >= 90 && c <= 97 {
			bg = false
			code = int32(c - 90)
			colorMode = render.ColorBit3Bright
		} else if c >= 100 && c <= 107 {
			bg = true
			code = int32(c - 100)
			colorMode = render.ColorBit3Bright
		} else {
			log.Printf("Unrecognized SGR code: %v", parameterCode)
		}

		color := render.Color{ColorMode: colorMode, Code: code}
		if bg {
			v.Cursor.Style.Bg = color
		} else {
			v.Cursor.Style.Fg = color
		}
	}
}
