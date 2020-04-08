package main

import "log"

func findClosestBorderForCoord(path Path, x, y int) Path {
	switch c := path.getContainer().(type) {
	case *Split:
		// NB: this works for Split(Term) but not Split(Split(Term, Term...))
		if len(c.elements) == 1 {
			return path
		}

		for idx, node := range c.elements {
			r := node.contents.getRenderRect()

			vertValid := c.verticallyStacked && y > r.y && y <= r.y+r.h+1
			horizValid := !c.verticallyStacked && x > r.x && x <= r.x+r.w+1
			if vertValid || horizValid {
				newPath := append(path, idx)
				return findClosestBorderForCoord(newPath, x, y)
			}
		}
	case *Pane:
		return path
	default:
		log.Println("Unexpected type! ", c)
	}

	return nil
}
