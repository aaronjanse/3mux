package vterm

/* actions.go implements functions for common operations upon the vterm screen */

func (v *VTerm) scrollDown(numLines int) {
	if !v.usingAltScreen {
		v.scrollback = append(v.scrollback, v.screen[v.scrollingRegion.top:v.scrollingRegion.top+numLines]...)
	}

	newLines := make([][]Char, numLines)

	v.screen = append(append(append(
		v.screen[:v.scrollingRegion.top],
		v.screen[v.scrollingRegion.top+numLines:v.scrollingRegion.bottom+1]...),
		newLines...),
		v.screen[v.scrollingRegion.bottom+1:]...)

	v.RedrawWindow()
}
