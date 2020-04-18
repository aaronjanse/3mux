package vterm

import "github.com/aaronjanse/3mux/ecma48"

func (v *VTerm) ProcessStdin(in ecma48.Output) []byte {
	return []byte(string(in.Raw))
}
