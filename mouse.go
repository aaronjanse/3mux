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

			if (c.verticallyStacked && r.y >= y) || (!c.verticallyStacked && r.x >= x) {
				newPath := append(path, idx-1)
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
