package main

import (
	"time"

	"github.com/aaronjanse/i3-tmux/keypress"
	"github.com/aaronjanse/i3-tmux/render"
)

func doDemo() {
	time.Sleep(500 * time.Millisecond)

	for _, r := range "vim RE" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(150 * time.Millisecond)
	}

	for _, r := range "ADME.md" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
	}

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(750 * time.Millisecond)
	handleInput(keypress.AltChar{Char: 'N'}, []byte{})

	time.Sleep(400 * time.Millisecond)

	for _, r := range "cat ma" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(150 * time.Millisecond)
	}
	for _, r := range "in.go" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
	}

	time.Sleep(400 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(1500 * time.Millisecond)
	handleInput(keypress.ScrollUp{}, []byte{})
	time.Sleep(100 * time.Millisecond)
	handleInput(keypress.ScrollUp{}, []byte{})
	time.Sleep(75 * time.Millisecond)
	handleInput(keypress.ScrollUp{}, []byte{})
	time.Sleep(50 * time.Millisecond)
	handleInput(keypress.ScrollUp{}, []byte{})
	time.Sleep(40 * time.Millisecond)
	handleInput(keypress.ScrollUp{}, []byte{})

	for i := 0; i < 10; i++ {
		time.Sleep(30 * time.Millisecond)
		handleInput(keypress.ScrollUp{}, []byte{})
	}

	time.Sleep(750 * time.Millisecond)

	handleInput(keypress.ScrollDown{}, []byte{})
	time.Sleep(75 * time.Millisecond)
	handleInput(keypress.ScrollDown{}, []byte{})
	time.Sleep(35 * time.Millisecond)
	handleInput(keypress.ScrollDown{}, []byte{})
	time.Sleep(35 * time.Millisecond)
	handleInput(keypress.ScrollDown{}, []byte{})
	time.Sleep(35 * time.Millisecond)
	handleInput(keypress.ScrollDown{}, []byte{})
	time.Sleep(35 * time.Millisecond)
	handleInput(keypress.ScrollDown{}, []byte{})

	time.Sleep(300 * time.Millisecond)
	handleInput(keypress.AltChar{Char: 'N'}, []byte{})

	time.Sleep(1000 * time.Millisecond)

	for _, r := range "cloc ." {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(750 * time.Millisecond)
	handleInput(keypress.AltShiftArrow{Direction: keypress.Down}, []byte{})

	time.Sleep(1000 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	for _, r := range "\ncat sp" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(150 * time.Millisecond)
	}
	for _, r := range "lit.go" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
	}

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(300 * time.Millisecond)
	handleInput(keypress.AltChar{Char: '/'}, []byte{})

	for _, r := range "re\n" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)
	handleInput(keypress.Character{Char: 'N'}, []byte{byte('N')})

	time.Sleep(300 * time.Millisecond)
	handleInput(keypress.Character{Char: 'N'}, []byte{byte('N')})

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: 'N'}, []byte{byte('N')})

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: 'N'}, []byte{byte('N')})

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: 'N'}, []byte{byte('N')})

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: 'n'}, []byte{byte('n')})

	time.Sleep(1750 * time.Millisecond)
	handleInput(keypress.AltShiftChar{Char: 'Q'}, []byte{})

	time.Sleep(750 * time.Millisecond)
	handleInput(keypress.AltArrow{Direction: keypress.Left}, []byte{})

	time.Sleep(500 * time.Millisecond)

	for _, r := range "\x1b/Feat\n" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(300 * time.Millisecond)
	}

	time.Sleep(750 * time.Millisecond)
	handleInput(keypress.AltChar{Char: 'F'}, []byte{})

	time.Sleep(500 * time.Millisecond)

	root.setPause(true)

	time.Sleep(200 * time.Millisecond)

	scr := getSelection().getContainer().(*Pane).vterm.Screen

	copiedText := ""

	for i := 9; i <= 17; i++ {
		w := termW
		if i == 17 {
			w = 20
		}
		for x := 0; x < w; x++ {
			r := scr[i][x].Rune
			if r != 0 && x < 30 {
				copiedText += string(r)
			}
			fg := scr[i][x].Style.Fg
			if fg.ColorMode == render.ColorNone || x > 10 {
				fg.ColorMode = render.ColorBit3Bright
				fg.Code = 7
			}
			bg := scr[i][x].Style.Bg
			if bg.ColorMode == render.ColorNone {
				bg.ColorMode = render.ColorBit3Normal
				bg.Code = 0
			}
			ch := render.PositionedChar{
				Rune: r,
				Cursor: render.Cursor{
					X: x,
					Y: i,
					Style: render.Style{
						Bg: fg,
						Fg: bg,
					},
				},
			}
			renderer.ForceHandleCh(ch)
		}
		time.Sleep(15 * time.Millisecond)
		copiedText += "\n"
	}

	time.Sleep(500 * time.Millisecond)

	getSelection().getContainer().(*Pane).vterm.ChangePause <- false
	getSelection().getContainer().(*Pane).vterm.RedrawWindow()

	// time.Sleep(5000000 * time.Millisecond)

	time.Sleep(1500 * time.Millisecond)
	handleInput(keypress.AltChar{Char: 'F'}, []byte{})

	time.Sleep(750 * time.Millisecond)

	for _, r := range "\x1b:q" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(300 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	for _, r := range "cmatrix" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(150 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(750 * time.Millisecond)
	handleInput(keypress.AltArrow{Direction: keypress.Right}, []byte{})

	for _, r := range "vim /tmp/x" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(150 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(500 * time.Millisecond)
	handleInput(keypress.Character{Char: 'i'}, []byte{byte('i')})

	time.Sleep(500 * time.Millisecond)

	for _, r := range "3mux is a multiplexer inspired by i3." {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(13 * time.Millisecond)
	}

	time.Sleep(1000 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})
	time.Sleep(300 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(500 * time.Millisecond)

	for _, r := range copiedText {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
	}

	time.Sleep(1000 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	for _, r := range "See more at https://github.com/aaronjanse/3mux." {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
		time.Sleep(13 * time.Millisecond)
	}

	time.Sleep(400 * time.Millisecond)
	handleInput(keypress.Character{Char: '\n'}, []byte{byte('\n')})

	time.Sleep(7000 * time.Millisecond)

	for _, r := range "\x1b:q!\n" {
		handleInput(keypress.Character{Char: r}, []byte{byte(r)})
	}

	time.Sleep(300 * time.Millisecond)
	shutdownNow()
}
