package main

import (
	"strconv"
	"strings"
)

func handleMouseDown(data []byte) Path {
	code := string(data[5:])
	parts := strings.Split(code, ";")
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(strings.TrimSuffix(parts[1], "M"))

	// are we clicking a border? if so, which one?
	path := findClosestBorderForCoord([]int{}, x, y)
	pane := path.getContainer()
	r := pane.getRenderRect()

	if x == r.x+r.w+1 || y == r.y+r.h+1 {
		return path
	}

	return nil
}

func findClosestBorderForCoord(path Path, x, y int) Path {
	switch c := path.getContainer().(type) {
	case *Split:
		// NB: this works for Split(Term) but not Split(Split(Term, Term...))
		if len(c.elements) == 1 {
			return path
		}

		for idx, node := range c.elements {
			r := node.contents.getRenderRect()

			if (c.verticallyStacked && r.y >= y) || (!c.verticallyStacked && r.x >= x) {
				newPath := append(path, idx-1)
				return findClosestBorderForCoord(newPath, x, y)
			}
		}
	case *Pane:
		return path
	}

	return nil
}
