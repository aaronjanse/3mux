package wm

func (u *Universe) simplify() {
	for _, w := range u.workspaces {
		w.contents.simplify()
	}
}

// stuff like h(h(x), y) -> h(x, y)
func (s *split) simplify() {
	if len(s.elements) == 1 {
		switch child := (*s).elements[0].contents.(type) {
		case *split:
			child.simplify()
			s.verticallyStacked = child.verticallyStacked
			s.elements = child.elements
			s.selectionIdx = child.selectionIdx
		}
	} else {
		newElements := []SizedNode{}
		selectionIdx := (*s).selectionIdx
		for idx, n := range (*s).elements {
			switch child := n.contents.(type) {
			case *split:
				child.simplify()

				if len(child.elements) > 0 {
					if child.verticallyStacked == s.verticallyStacked {
						for j := range child.elements {
							child.elements[j].size *= n.size
						}
						newElements = append(newElements, child.elements...)
						if idx == s.selectionIdx {
							selectionIdx += child.selectionIdx
						} else if idx < s.selectionIdx {
							selectionIdx += len(child.elements) - 1
						}
					} else {
						newElements = append(newElements, n)
					}
				}
			case Node:
				newElements = append(newElements, n)
			}
		}
		s.elements = newElements
		s.selectionIdx = selectionIdx
		if s.selectionIdx > len(s.elements)-1 {
			s.selectionIdx = len(s.elements) - 1
		}

		var totalSize float32
		for _, e := range s.elements {
			totalSize += e.size
		}
		scale := 1 / totalSize
		for i := range s.elements {
			s.elements[i].size *= scale
		}
	}

	for _, n := range s.elements {
		switch child := n.contents.(type) {
		case *split:
			child.simplify()
		}
	}
}
