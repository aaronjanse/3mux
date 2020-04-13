package wm

import (
	"fmt"
)

// A Workspace is a desktop
type Workspace struct {
	contents     *Split
	doFullscreen bool
}

func (k *Workspace) Serialize() string {
	return fmt.Sprintf("Workspace(%s)", k.contents.Serialize())
}

func (k *Workspace) Reshape(x, y, w, h int) {
	// if s.doFullscreen {
	// 	getSelection().getContainer().Reshape(x, y, w, h)
	// } else {
	k.contents.Reshape(x, y, w, h)
	// }
}

func (k *Workspace) Kill() {
	k.contents.Kill()
}

func (k *Workspace) AddPane() {
	k.contents.AddPane()
}
