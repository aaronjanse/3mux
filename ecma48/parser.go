/*
Package ecma38 is super cool and in need of documentation!

All coordinates are relative to (0, 0) in the top left corner of the terminal.
*/

package ecma48

import (
	"bufio"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// Parser maintains state. Create one with NewParser() then start the parsing via Run()
type Parser struct {
	out chan<- Output

	state

	private      rune
	intermediate string
	params       string
	final        rune

	data []rune

	// RuneCounter is useful for detecting if the processer is lagging
	RuneCounter uint64
}

// NewParser creates a parser to be used for Parse()
func NewParser() *Parser {
	return &Parser{
		state:        stateGround,
		private:      0,
		intermediate: "",
		params:       "",
		final:        0,
		data:         []rune{},
		RuneCounter:  0,
	}
}

// Parse starts the parsing process, reading from input, parsing, then sending to output
func (p *Parser) Parse(input *bufio.Reader, output chan<- Output) error {
	p.out = output
	for {
		r, _, err := input.ReadRune()
		if err != nil {
			return err
		}

		p.data = append(p.data, r)

		atomic.AddUint64(&p.RuneCounter, 1)
		p.anywhere(r)
	}
}

func (p *Parser) wrap(x Parsed) Output {
	output := Output{
		Raw:    p.data,
		Parsed: x,
	}

	p.data = []rune{}

	return output
}

type state int

const (
	stateGround = iota
	stateEscape
	stateCsiEntry
	stateCsiParam
	stateOscString
)

func (p *Parser) anywhere(r rune) {
	switch r {
	case 0x00:
	case 0x1B:
		p.doClear()
		p.state = stateEscape
	case 0x8D: // Reverse Index
		p.out <- p.wrap(RI{})
		p.state = stateGround
	case 0x9B:
		p.doClear()
		p.state = stateCsiEntry
	case 0x9C:
		p.state = stateGround
	case 0x9D:
		p.state = stateOscString
	default:
		switch p.state {
		case stateGround:
			p.stateGround(r)
		case stateEscape:
			p.stateEscape(r)
		case stateCsiEntry:
			p.stateCsiEntry(r)
		case stateCsiParam:
			p.stateCsiParam(r)
		case stateOscString:
			p.stateOscString(r)
		default:
			log.Printf("? STATE %d", p.state)
			p.state = stateGround
		}
	}
}

func (p *Parser) stateGround(r rune) {
	switch {
	case '\b' == r:
		p.out <- p.wrap(Backspace{})
	case '\n' == r:
		p.out <- p.wrap(Newline{})
	case '\r' == r:
		p.out <- p.wrap(CarriageReturn{})
	case '\t' == r:
		p.out <- p.wrap(Tab{})
	case unicode.IsPrint(r):
		p.out <- p.wrap(Char{
			Rune:   r,
			IsWide: runewidth.RuneWidth(r) > 1,
		})
	default:
		log.Printf("? GROUND %s (%d)", string(r), r)
	}
}

func (p *Parser) stateEscape(r rune) {
	switch {
	case strings.Contains("DEHMNOPVWXZ[\\]^_", string(r)):
		p.anywhere(r + 0x40)
	case 0x30 <= r && r <= 0x4F || 0x51 <= r && r <= 0x57:
		// TODO: p.dispatchEsc()
		p.state = stateGround
	default:
		// log.Printf("? ESC %s", string(r))
	}
}

func (p *Parser) stateCsiEntry(r rune) {
	switch {
	case 0x30 <= r && r <= 0x39 || r == 0x3B:
		p.params += string(r)
		p.state = stateCsiParam
	case 0x3C <= r && r <= 0x3F:
		p.intermediate += string(r)
		p.state = stateCsiParam
	case 0x40 <= r && r <= 0x7E:
		p.final = r
		p.dispatchCsi()
		p.state = stateGround
	}
}

func (p *Parser) stateCsiParam(r rune) {
	switch {
	case 0x30 <= r && r <= 0x39 || r == 0x3b:
		p.params += string(r)
	case 0x40 <= r && r <= 0x7E:
		p.final = r
		p.dispatchCsi()
		p.state = stateGround
	}
}

func (p *Parser) stateOscString(r rune) {
	switch {
	case 0x07 == r: // FIXME: this is weird
		p.state = stateGround
	}
}

func (p *Parser) doClear() {
	p.private = 0
	p.intermediate = ""
	p.params = ""
	p.final = 0
}

func (p *Parser) dispatchCsi() {
	switch p.intermediate {
	case "?":
		switch p.final {
		case 'h': // DECSET
			i, err := strconv.Atoi(p.params)
			if err == nil {
				p.out <- p.wrap(PrivateDEC{On: true, Code: i})
			} else {
				p.out <- p.wrap(Unrecognized("DECSET"))
			}
		case 'l': // DECRST
			i, err := strconv.Atoi(p.params)
			if err == nil {
				p.out <- p.wrap(PrivateDEC{On: false, Code: i})
			} else {
				p.out <- p.wrap(Unrecognized("DECRST"))
			}
		default:
			p.out <- p.wrap(Unrecognized("DEC Private Mode"))
		}
	case "":
		switch p.final {
		case 'A', 'B', 'C', 'D':
			seq := parseSemicolonNumSeq(p.params, 1)
			n := seq[0]

			var dir Direction
			switch p.final {
			case 'A':
				dir = Up
			case 'B':
				dir = Down
			case 'C':
				dir = Right
			case 'D':
				dir = Left
			}

			if n > 0 {
				if len(seq) > 1 {
					p.out <- p.wrap(CursorMovement{
						N: n, Direction: dir,

						Shift: (seq[1]-1)&0b001 > 0,
						Alt:   (seq[1]-1)&0b010 > 0,
						Ctrl:  (seq[1]-1)&0b100 > 0,
					})
				} else {
					p.out <- p.wrap(CursorMovement{N: n, Direction: dir})
				}
			}
		case 'd': // Vertical Line Position Absolute (VPA)
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(VPA{seq[0] - 1})
		case 'E': // Cursor Next Line
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(CNL{YDiff: uint(seq[0])}) // FIXME bad error handling
		case 'F': // Cursor Previous Line
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(CPL{YDiff: uint(seq[0])}) // FIXME bad error handling
		case 'G': // Cursor Horizontal Absolute
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(CHA{X: seq[0] - 1})
		case 'H', 'f': // Cursor Position
			seq := parseSemicolonNumSeq(p.params, 1)
			if len(seq) > 1 {
				p.out <- p.wrap(CUP{Y: seq[0] - 1, X: seq[1] - 1})
			} else {
				p.out <- p.wrap(CUP{Y: seq[0] - 1, X: 0})
			}
		case 'J': // Erase in Display
			seq := parseSemicolonNumSeq(p.params, 0)
			p.out <- p.wrap(ED{Directive: seq[0]})
		case 'K': // Erase in Line
			seq := parseSemicolonNumSeq(p.params, 0)
			p.out <- p.wrap(EL{Directive: seq[0]})
		case 'L': // Insert Lines; https://vt100.net/docs/vt510-rm/IL.html
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(IL{N: seq[0]})
		case 'M': // Delete Lines; https://vt100.net/docs/vt510-rm/DL.html
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(DL{N: seq[0]})
		case 'n': // Device Status Report (TODO)
			// seq := parseSemicolonNumSeq(p.params, 0)
			// switch seq[0] {
			// case 6:
			// 	response := fmt.Sprintf("\x1b[%d;%dR", v.Cursor.Y+1, v.Cursor.X+1)
			// 	for _, r := range response {
			// 		v.out <- r
			// 	}
			// default:
			// 	log.Println("Unrecognized DSR code", seq)
			// }
		case 'r': // Set Scrolling Region
			seq := parseSemicolonNumSeq(p.params, 1)
			if len(seq) > 1 {
				p.out <- p.wrap(DECSTBM{Top: seq[0] - 1, Bottom: seq[1] - 1})
			} else {
				p.out <- p.wrap(DECSTBM{Top: seq[0] - 1, Bottom: -1})
			}
		case 'S': // Scroll Up; new lines added to bottom
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(SU{N: uint(seq[0])}) // FIXME
		case 'T': // Scroll Down; new lines added to top
			seq := parseSemicolonNumSeq(p.params, 1)
			p.out <- p.wrap(SD{N: uint(seq[0])}) // FIXME
		// case 't': // Window Manipulation
		// 	// TODO
		case 'm': // Select Graphic Rendition
			p.handleSGR(p.params)
		case 's': // Save Cursor Position
			p.out <- p.wrap(SCOSC{})
		case 'u': // Restore Cursor Positon
			p.out <- p.wrap(SCORC{})
		default:
			log.Printf("? CSI %s %s", p.params, string(p.final))
		}
	}
}

func (p *Parser) handleSGR(parameterCode string) {
	seq := parseSemicolonNumSeq(parameterCode, 0)

	if parameterCode == "39;49" {
		p.out <- p.wrap(StyleForeground(Color{ColorMode: ColorNone}))
		p.data = []rune{}
		p.out <- p.wrap(StyleBackground(Color{ColorMode: ColorNone}))
		return
	}

	for {
		if len(seq) == 0 {
			break
		}

		c := seq[0]

		switch c {
		case 0:
			p.out <- p.wrap(StyleReset{})
			seq = seq[1:]
		case 1:
			p.out <- p.wrap(StyleBold(true))
			seq = seq[1:]
		case 2:
			p.out <- p.wrap(StyleFaint(true))
			seq = seq[1:]
		case 3:
			p.out <- p.wrap(StyleItalic(true))
			seq = seq[1:]
		case 4:
			p.out <- p.wrap(StyleUnderline(true))
			seq = seq[1:]
		case 5: // slow blink
			seq = seq[1:]
		case 6: // rapid blink
			seq = seq[1:]
		case 7: // swap foreground & background; see case 27
			seq = seq[1:]
		case 8:
			p.out <- p.wrap(StyleConceal(true))
			seq = seq[1:]
		case 9:
			p.out <- p.wrap(StyleCrossedOut(true))
			seq = seq[1:]
		case 10: // primary/default font
			p.out <- p.wrap(StyleUnderline(false))
			seq = seq[1:]
		case 22:
			p.out <- p.wrap(StyleBold(false))
			p.out <- p.wrap(StyleFaint(false))
			seq = seq[1:]
		case 23:
			p.out <- p.wrap(StyleItalic(false))
			seq = seq[1:]
		case 24:
			p.out <- p.wrap(StyleUnderline(false))
			seq = seq[1:]
		case 25: // blink off
			seq = seq[1:]
		case 27: // inverse off; see case 7
			seq = seq[1:]
		case 28:
			p.out <- p.wrap(StyleConceal(false))
			seq = seq[1:]
		case 29:
			p.out <- p.wrap(StyleCrossedOut(false))
			seq = seq[1:]

		case 38: // set foreground color
			if seq[1] == 5 {
				p.out <- p.wrap(StyleForeground(Color{
					ColorMode: ColorBit8,
					Code:      int32(seq[2]),
				}))
				seq = seq[3:]
			} else if seq[1] == 2 {
				p.out <- p.wrap(StyleForeground(Color{
					ColorMode: ColorBit24,
					Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
				}))
				seq = seq[5:]
			}
		case 39: // default foreground color
			p.out <- p.wrap(StyleForeground(Color{ColorMode: ColorNone}))
			seq = seq[1:]
		case 48: // set background color
			if seq[1] == 5 {
				p.out <- p.wrap(StyleBackground(Color{
					ColorMode: ColorBit8,
					Code:      int32(seq[2]),
				}))
				seq = seq[3:]
			} else if seq[1] == 2 {
				p.out <- p.wrap(StyleBackground(Color{
					ColorMode: ColorBit24,
					Code:      int32(seq[2]<<16 + seq[3]<<8 + seq[4]),
				}))
				seq = seq[5:]
			}
		case 49: // default background color
			p.out <- p.wrap(StyleBackground(Color{ColorMode: ColorNone}))
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
					seq = seq[2:]
				} else {
					colorMode = ColorBit3Normal
					seq = seq[1:]
				}
			} else if c >= 40 && c <= 47 {
				bg = true
				code = int32(c - 40)
				if len(seq) > 1 && seq[1] == 1 {
					colorMode = ColorBit3Bright
					seq = seq[2:]
				} else {
					colorMode = ColorBit3Normal
					seq = seq[1:]
				}
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
				p.out <- p.wrap(StyleBackground(color))
			} else {
				p.out <- p.wrap(StyleForeground(color))
			}
		}
	}
}

// parseSemicolonNumSeq parses a series of numbers separated by semicolons, replacing empty values with the given default value
// FIXME: this function is an unclean way to parse parameters, espcially when it comes to default values
func parseSemicolonNumSeq(s string, d int) []int {
	s = strings.TrimSpace(s)

	if s == "" {
		return []int{d}
	}

	parts := strings.Split(s, ";")

	out := []int{}
	for _, part := range parts {
		if part == "" {
			out = append(out, d)
		} else {
			num, err := strconv.ParseInt(part, 10, 32)
			if err != nil {
				continue // fixme?
			}

			out = append(out, int(num))
		}
	}
	return out
}
