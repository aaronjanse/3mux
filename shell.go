package main

import (
	"fmt"
	"math/rand"

	"github.com/aaronduino/i3-tmux/ansi"
)

// Markup is text that may or may not contain special escape codes such as ANSI CSI Sequences
type Markup string

// A Term is the fundamental leaf unit of screen space
type Term struct {
	id int

	// buffer is pure; it is not tailted by rewriteForOrigin
	buffer     Markup
	renderRect Rect

	selected bool
}

func newTerm(selected bool) *Term {
	return &Term{
		id:       rand.Intn(10),
		buffer:   "xxxxx",
		selected: selected,
	}
}

func (t *Term) handleStdout(text Markup) {
	t.buffer += text

	// TODO: truncate buffer if necessary

	transformed := text.rewrite(t.renderRect, t.selected)
	fmt.Print(transformed)
}

// rewrite CSI codes for an origin at the given coordinates
func (m Markup) rewrite(r Rect, selected bool) Markup {
	// FIXME: actually rewrite CSI codes

	out := Markup(ansi.MoveTo(r.x, r.y)) + m

	if !selected {
		out = "\033[2m" + out + "\033[m"
	} else {
		out = "\033[4m" + out + "\033[m"
	}

	return out
}
