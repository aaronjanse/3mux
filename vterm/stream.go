package vterm

import (
	"log"
	"sync/atomic"
	"time"
)

// pullRune returns the next byte in the input stream
func (v *VTerm) pullRune() (rune, bool) {
	v.internalByteCounter++

	lag := atomic.LoadUint64(v.shellByteCounter) - v.internalByteCounter
	if lag > uint64(v.w*v.h*4) {
		v.useSlowRefresh()
	} else {
		v.useFastRefresh()
	}

	for {
		select {
		case r, ok := <-v.in:
			if r != 0 {
				if v.DebugSlowMode {
					log.Printf("rune: %v (%s)", r, string(r))
					time.Sleep(100 * time.Millisecond)
				}
				return r, ok
			}
		case p := <-v.ChangePause:
			for {
				v.IsPaused = p
				if !p {
					break
				}
				p = <-v.ChangePause
			}
		}
	}
}
