package keypress

// Esc is for when the escape key is pressed
type Esc struct{}

// Enter is the same thing as \n or return
type Enter struct{}

// Character is for when a single non-modifier key is pressed.
// It has a long name to prevent confusion with Char used in other structs
type Character struct {
	Char rune
}

// CtrlChar is for ctrl+[character]
type CtrlChar struct {
	Char rune
}

// AltChar is for alt+[character]
type AltChar struct {
	Char rune
}

// AltShiftChar is for alt+shift+[character]
type AltShiftChar struct {
	Char rune
}

// Arrow is for the arrow keys
type Arrow struct {
	Direction
}

// ShiftArrow is for shift+[arrow key]
type ShiftArrow struct {
	Direction
}

// AltArrow is for alt+[arrow key]
type AltArrow struct {
	Direction
}

// AltShiftArrow is for alt+shift+[arrow key]
type AltShiftArrow struct {
	Direction
}

// CtrlArrow is for ctrl+[arrow key]
type CtrlArrow struct {
	Direction
}

// ScrollUp moves the viewport up, and the screen contents down
type ScrollUp struct{}

// ScrollDown moves the viewport down, and the screen contents up
type ScrollDown struct{}

// MouseMove is for when the terminal reports mouse movement
type MouseMove struct {
	mouseDown bool
	x, y      int
}

// MouseDown is for when the primary mouse button is pressed down
type MouseDown struct {
	X, Y int
}

// MouseUp is for when the primary mouse button is no longer being pressed
type MouseUp struct {
	X, Y int
}

// Resize is for host terminal resize eventd
type Resize struct {
	W, H int
}
