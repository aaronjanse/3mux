package main

import (
	"fmt"
	"os"

	tsm "github.com/emersion/go-tsm"
)

func main() {
	screen := tsm.NewScreen()
	screen.Reset()
	screen.Resize(2, 2)

	vte := tsm.NewVTE(screen, func(b []byte) {
		fmt.Fprintf(os.Stderr, "aaaa")
	})

	// vte.Input([]byte("Hello"))
	vte.Input([]byte("\033[1ma\033[0m"))

	screen.Draw(func(id uint32, s string, width, posx, posy uint, attr *tsm.ScreenAttr, age uint32) bool {
		// fmt.Printf("\033[%d;%dH%s", posy, posx+1, s)
		fmt.Println(age, id)
		// fmt.Print(id)
		// fmt.Println(attr)
		return false
	})
	vte.Input([]byte("\033[1mb\033[0m"))

	fmt.Println()

	screen.Draw(func(id uint32, s string, width, posx, posy uint, attr *tsm.ScreenAttr, age uint32) bool {
		// fmt.Printf("\033[%d;%dH%s", posy, posx+1, s)
		fmt.Println(age, id)
		// fmt.Print(id)
		// fmt.Println(attr)
		return false
	})

	// screen.Reset()
	// screen.Resize(100, 8)
	// a := &tsm.ScreenAttr{0, 0, 255, 255, 255, 0, 0, 0, false, false, false, false, false}
	// for _, r := range []rune("\033[1mHello there\nWorld\033[0m") {
	// 	if r == '\n' {
	// 		screen.Newline()
	// 		screen.MoveDown(1, false)
	// 	}
	// 	screen.Write(r, a)
	// 	// fmt.Println()
	// }

	// screen.Draw(func(id uint32, s string, width, posx, posy uint, attr *tsm.ScreenAttr, age uint32) bool {
	// 	fmt.Printf("\033[%d;%dH%s", posy, posx+1, s)
	// 	// fmt.Print(age)
	// 	// fmt.Println()
	// 	return true
	// })
}
