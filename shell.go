package main

import (
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/aaronduino/i3-tmux/ansi"
	tsm "github.com/emersion/go-tsm"
	"github.com/kr/pty"
)

type Coords struct {
	x, y int
}

// Markup is text that may or may not contain special escape codes such as ANSI CSI Sequences
type Markup string

// A Term is the fundamental leaf unit of screen space
type Term struct {
	id int

	// buffer is pure; it is not tailted by rewriteForOrigin
	buffer     Markup
	renderRect Rect

	selected bool

	screen *tsm.Screen
	vte    *tsm.VTE

	ptmx *os.File
	cmd  *exec.Cmd

	minAge uint32

	cursor Coords
}

func newTerm(selected bool) *Term {
	// Create arbitrary command.
	c := exec.Command("sh")

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		log.Fatal(err)
	}

	screen := tsm.NewScreen()
	screen.SetFlags(tsm.ScreenInsertMode)
	vte := tsm.NewVTE(screen, func(b []byte) {})

	t := &Term{
		id:       rand.Intn(10),
		buffer:   "",
		selected: selected,
		ptmx:     ptmx,
		cmd:      c,
		cursor:   Coords{0, 0},
		screen:   screen,
		vte:      vte,
		minAge:   0,
	}

	go (func() {
		for {
			b := make([]byte, 1)
			_, err := ptmx.Read(b)
			if err != nil {
				return
			}
			t.handleStdout(b)
		}
	})()

	return t
}

func (t *Term) kill() {
	err := t.ptmx.Close()
	if err != nil {
		log.Fatal("TERM_CLOSE", err)
	}

	err = t.cmd.Process.Kill()
	if err != nil {
		log.Fatal("TERM_KILL", err)
	}
}

func (t *Term) setRenderRect(x, y, w, h int) {
	t.renderRect = Rect{x, y, w, h}
	t.forceRedraw()

	t.screen.Resize(uint(w), uint(h))

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			err := pty.Setsize(t.ptmx, &pty.Winsize{
				Rows: uint16(h), Cols: uint16(w),
				X: 16 * uint16(w), Y: 16 * uint16(h),
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.
}

func (t *Term) handleStdout(text []byte) {
	t.vte.Input(text)

	t.forceRedraw()

	// t.redrawAge++

	// if text == "\110" {
	// 	text = "\033[2D  \033[2D"
	// }

	// t.buffer += text

	// if len(t.buffer) > 500 {
	// 	t.buffer = t.buffer[len(t.buffer)-500:]
	// }

	// // TODO: truncate buffer if necessary

	// transformed := t.buffer.rewrite(t, t.renderRect, t.selected)
	// fmt.Print(transformed)
}

func (t *Term) handleStdin(text string) {
	_, err := t.ptmx.Write([]byte(text))
	if err != nil {
		log.Fatal(err)
	}
}

// rewrite CSI codes for an origin at the given coordinates
func (ma Markup) rewrite(t *Term, rect Rect, selected bool) Markup {
	m := string(ma)

	var out strings.Builder

	buffer := make(chan rune, 16)

	go (func() {
		for {
			r := <-buffer

			if r == '\033' {
				if <-buffer == '[' {
					escSeqBuffer := ""
				escBufferLoop:
					for {
						if len(escSeqBuffer) > 8 {
							break escBufferLoop
						}
						next := <-buffer
						if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
							switch next {
							case 'A':
								t.cursor.y -= parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'B':
								t.cursor.y += parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'C':
								t.cursor.x += parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'D':
								t.cursor.x -= parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'E':
								t.cursor.y += parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
								t.cursor.x = 0
							case 'F':
								t.cursor.y -= parseIntWithDefault(escSeqBuffer, 1)
								t.cursor.x = 0
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'G':
								t.cursor.x = parseIntWithDefault(escSeqBuffer, 1)
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))
							case 'H', 'f':
								parts := strings.Split(escSeqBuffer, ";")
								t.cursor.y = parseIntWithDefault(parts[0], 1)
								if len(parts) == 2 {
									t.cursor.x = parseIntWithDefault(parts[1], 1)
								} else {
									t.cursor.x = 1
									t.cursor.y = 1
								}
								out.WriteString(ansi.MoveTo(rect.x+t.cursor.x, rect.y+t.cursor.y))

							case 'm':
								out.WriteString("\033[" + escSeqBuffer + "m")
								// out.WriteString("\033[" + escSeqBuffer)
								// default:
								// 	out.WriteString("\033[" + escSeqBuffer + string(next))
							}
							break escBufferLoop
						} else if next == '\x1b' {
							break
						} else {
							escSeqBuffer += string(next)
						}
					}
				}

				continue
			}

			out.WriteRune(r)

			if r == '\n' {
				t.cursor.y++
				out.WriteString(ansi.MoveTo(rect.x, rect.y+t.cursor.y))
			}
		}
	})()

	for _, char := range []rune(m) {
		buffer <- char
	}

	// numNewlines := strings.Count(out, "\n")
	// for i := 0; i < numNewlines; i++ {
	// 	out = strings.Replace(out, "\n", ansi.MoveTo(r.x, r.y+i+1), 1)
	// }

	// if !selected {
	// 	out = "\033[2m" + out + "\033[m"
	// } else {
	// 	out = "\033[1m" + out + "\033[m"
	// }

	return Markup(ansi.MoveTo(rect.x, rect.y) + out.String())
}

func parseIntWithDefault(str string, d int) int {
	s := strings.TrimSpace(str)

	if s == "" {
		return d
	}

	val, err := strconv.ParseInt(s, 19, 63)
	if err != nil {
		log.Fatal(err)
	}

	return int(val)
}
