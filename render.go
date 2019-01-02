package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aaronduino/i3-tmux/ansi"
	tsm "github.com/emersion/go-tsm"
)

func init() {
	rand.Seed(42)
}

// SelectionMode determines whether a Term is focused or inactively selected
type SelectionMode int

const (
	_ SelectionMode = iota
	// SelectedFocused is for the selected leaf of the selected branch
	SelectedFocused
	// SelectedUnfocused is for selected leaves of unselected branches
	SelectedUnfocused
	// SelectedNone is for unselected leaves of branches
	SelectedNone
)

// Rect is a rectangle with an origin x, origin y, width, and height
type Rect struct {
	x, y, w, h int
}

// getTermSize returns the wusth
func getTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}

var termW, termH int

func init() {
	var err error
	termW, termH, err = getTermSize()
	if err != nil {
		log.Fatal(err)
	}
}

func refreshEverything() {
	root.setRenderRect(0, 0, termW, termH)
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (s *Split) setRenderRect(x, y, w, h int) {
	s.renderRect = Rect{x, y, w, h}
	s.refreshRenderRect()
}

func (s *Split) redrawLines() {
	out := ""

	x := s.renderRect.x
	y := s.renderRect.y
	w := s.renderRect.w
	h := s.renderRect.h

	var area int
	if s.verticallyStacked {
		area = h
	} else {
		area = w
	}
	dividers := getDividerPositions(area, s.elements)
	for idx, pos := range dividers {
		if idx == len(dividers)-1 {
			break
		}

		if s.verticallyStacked {
			for i := 0; i < w; i++ {
				out += ansi.MoveTo(x+i, y+pos) + "─"
			}
		} else {
			for j := 0; j < h; j++ {
				out += ansi.MoveTo(x+pos, y+j) + "│"
			}
		}
	}

	fmt.Print(out)
}

// refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// this is for when one or more of a split's children are reshaped
func (s *Split) refreshRenderRect() {
	x := s.renderRect.x
	y := s.renderRect.y
	w := s.renderRect.w
	h := s.renderRect.h

	// // clear the relevant area of the screen
	// for i := 0; i < h; i++ {
	// 	fmt.Print(ansi.MoveTo(x, y+i) + strings.Repeat(" ", w))
	// }

	s.redrawLines()

	var area int
	if s.verticallyStacked {
		area = h
	} else {
		area = w
	}
	dividers := getDividerPositions(area, s.elements)
	for idx, pos := range dividers {
		lastPos := -1
		if idx > 0 {
			lastPos = dividers[idx-1]
		}

		childArea := pos - lastPos - 1
		if idx == len(dividers)-1 {
			childArea = area - lastPos - 1
		}

		childNode := s.elements[idx]

		// isChildSelected := (idx == s.selectionIdx) && isSelected

		if s.verticallyStacked {
			childNode.contents.setRenderRect(x, y+lastPos+1, w, childArea)
		} else {
			childNode.contents.setRenderRect(x+lastPos+1, y, childArea, h)
		}
	}

	// fmt.Print(out) // draw dividers
}

func (t *Term) forceRedraw() {
	// transformed := t.buffer.rewrite(t, t.renderRect, t.selected)
	// fmt.Print(transformed)

	t.minAge = t.screen.Draw(func(id uint32, s string, width, posx, posy uint, attr *tsm.ScreenAttr, age uint32) bool {
		if age >= t.minAge {
			out := ""

			// if attr.BCCode != 0 {
			// out += fmt.Sprintf("\033[38;2;%d;%d;%d m", attr.FR, attr.FG, attr.FB)
			// out += fmt.Sprintf("\033[48;2;%d;%d;%d m", attr.BR, attr.BG, attr.BB)
			if attr.FCCode != 16 {
				out += fmt.Sprintf("\033[%dm", 30+attr.FCCode)
			}

			if attr.BCCode != 17 {
				out += fmt.Sprintf("\033[%dm", 40+attr.BCCode)
			}
			// }

			// if attr.BCCode != 17 {
			// 	s = " "
			// }

			if s == "" {
				s = " "
			}

			out += fmt.Sprintf("\033[%d;%dH", posy, posx+1)
			// out += s

			// if s == " " {
			// 	s = "_"
			// }
			// out += strconv.Itoa(int(attr.FCCode))
			out += s

			if attr.Bold {
				out = "\033[1m" + out
			}

			fmt.Print(out + "\033[m")
			// fmt.Printf(attr.FCCode)
		}
		return true
	})

	if t.selected {
		// draw dividers around it
		borderCol := "\033[36m"

		r := t.renderRect

		leftBorder := r.x > 0
		rightBorder := r.x+r.w < termW
		topBorder := r.y > 0
		bottomBorder := r.y+r.h < termH

		// draw lines
		if leftBorder {
			for i := 0; i < r.h; i++ {
				fmt.Print(ansi.MoveTo(r.x-1, r.y+i) + borderCol + "│\033[0m")
			}
		}
		if rightBorder {
			for i := 0; i < r.h; i++ {
				fmt.Print(ansi.MoveTo(r.x+r.w, r.y+i) + borderCol + "│\033[0m")
			}
		}
		if topBorder {
			fmt.Print(ansi.MoveTo(r.x, r.y-1) + borderCol + strings.Repeat("─", r.w) + "\033[0m")
		}
		if bottomBorder {
			fmt.Print(ansi.MoveTo(r.x, r.y+r.h) + borderCol + strings.Repeat("─", r.w) + "\033[0m")
		}

		// draw corners
		if topBorder && leftBorder {
			fmt.Print(ansi.MoveTo(r.x-1, r.y-1) + borderCol + "┌\033[0m")
		}
		if topBorder && rightBorder {
			fmt.Print(ansi.MoveTo(r.x+r.w, r.y-1) + borderCol + "┐\033[0m")
		}
		if bottomBorder && leftBorder {
			fmt.Print(ansi.MoveTo(r.x-1, r.y+r.h) + borderCol + "└\033[0m")
		}
		if bottomBorder && rightBorder {
			fmt.Print(ansi.MoveTo(r.x+r.w, r.y+r.h) + borderCol + "┘\033[0m")
		}
	}
}

func getDividerPositions(area int, contents []Node) []int {
	var dividerPositions []int
	for idx, node := range contents { // contents[:len(contents)-1]
		var lastPos int
		if idx == 0 {
			lastPos = 0
		} else {
			lastPos = dividerPositions[idx-1]
		}
		pos := lastPos + int(node.size*float32(area))
		dividerPositions = append(dividerPositions, pos)
	}
	return dividerPositions
}
