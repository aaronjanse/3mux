package vterm

import (
	"time"
)

/*
https://github.com/tmux/tmux/issues/849#issuecomment-291828893

When we are lagging far behind, we should periodically refresh over a greater period of time than usual.

Otherwise, we should refresh as often as reasonably possible.
*/

func (v *VTerm) useSlowRefresh() {
	if v.usingSlowRefresh {
		return
	}

	v.usingSlowRefresh = true

	go func() {
		ticker := time.NewTicker(time.Millisecond * 1000)

		for range ticker.C {
			if !v.usingSlowRefresh || v.IsPaused {
				ticker.Stop()
				return
			}

			v.RedrawWindow()
		}
	}()
}

func (v *VTerm) useFastRefresh() {
	if !v.usingSlowRefresh {
		return
	}

	v.usingSlowRefresh = false
}
