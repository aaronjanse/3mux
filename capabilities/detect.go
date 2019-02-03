package capabilities

import (
	"os"
)

// Caps holds data about a terminal's capabilities
type Caps struct {
	ScrollingRegionTopBottom,
	ScrollingRegionLeftRight,
	_ bool
}

// Capabilities is a struct containing info on the host terminal's capabilities
var Capabilities Caps

func init() {
	switch os.Getenv("TERM") {
	case "xterm":
		Capabilities = Caps{
			ScrollingRegionTopBottom: true,
			ScrollingRegionLeftRight: true,
		}
	case "xterm-kitty":
		Capabilities = Caps{
			ScrollingRegionTopBottom: true,
			ScrollingRegionLeftRight: false,
		}
	default:
		Capabilities = Caps{
			ScrollingRegionTopBottom: true,
			ScrollingRegionLeftRight: true,
		}
	}
}
