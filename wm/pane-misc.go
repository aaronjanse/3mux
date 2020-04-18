package wm

func (u *Universe) ToggleSearch() {
	u.workspaces[u.selectionIdx].contents.ToggleSearch()
}

func (s *split) ToggleSearch() {
	s.elements[s.selectionIdx].contents.ToggleSearch()
}

func (u *Universe) ScrollUp() {
	u.workspaces[u.selectionIdx].contents.ScrollUp()
}
func (s *split) ScrollUp() {
	s.elements[s.selectionIdx].contents.ScrollUp()
}

func (u *Universe) ScrollDown() {
	u.workspaces[u.selectionIdx].contents.ScrollDown()
}
func (s *split) ScrollDown() {
	s.elements[s.selectionIdx].contents.ScrollDown()
}

func (u *Universe) HandleStdin(str string) {
	u.workspaces[u.selectionIdx].contents.HandleStdin(str)
}
func (s *split) HandleStdin(str string) {
	s.elements[s.selectionIdx].contents.HandleStdin(str)
}
