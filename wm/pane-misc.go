package wm

import "github.com/aaronjanse/3mux/ecma48"

func (u *Universe) ToggleSearch() {
	u.wmOpMutex.Lock()
	defer u.wmOpMutex.Unlock()

	u.workspaces[u.selectionIdx].contents.ToggleSearch()
}

func (s *split) ToggleSearch() {
	if len(s.elements) == 0 {
		return
	}
	s.elements[s.selectionIdx].contents.ToggleSearch()
}

func (u *Universe) ScrollUp() {
	u.workspaces[u.selectionIdx].contents.ScrollUp()
}
func (s *split) ScrollUp() {
	if len(s.elements) == 0 {
		return
	}
	s.elements[s.selectionIdx].contents.ScrollUp()
}

func (u *Universe) ScrollDown() {
	u.workspaces[u.selectionIdx].contents.ScrollDown()
}
func (s *split) ScrollDown() {
	s.elements[s.selectionIdx].contents.ScrollDown()
}

func (u *Universe) HandleStdin(in ecma48.Output) {
	u.workspaces[u.selectionIdx].contents.HandleStdin(in)
}
func (s *split) HandleStdin(in ecma48.Output) {
	s.elements[s.selectionIdx].contents.HandleStdin(in)
}
