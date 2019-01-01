package ansi

import "fmt"

// MoveTo returns an ANSI escape sequence to move the cursor to the given coordinates
func MoveTo(x, y int) string {
	return fmt.Sprintf("\033[%d;%dH", y+1, x+1)
}

// EraseToEOL returns an ANSI escape sequence to clear from the cursor to the end of the line
func EraseToEOL() string {
	return "\033[2K"
}
