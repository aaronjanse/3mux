package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kr/pty"
)

// Shell manages spawning, killing, and sending data to/from a shell subprocess (e.g. bash, sh, zsh)
type Shell struct {
	stdout chan<- rune
	ptmx   *os.File
	cmd    *exec.Cmd
}

func newShell(stdout chan<- rune) Shell {
	// cmd := exec.Command("bash", "/home/ajanse/Playground/i3-tmux/test.sh")
	cmd := exec.Command("zsh")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	// feed ptmx output to stdout channel
	go (func() {
		defer func() {
			if r := recover(); r != nil {
				if r.(error).Error() != "send on closed channel" {
					panic(r)
				}
			}
		}()

		defer func() {
			nowTime := time.Now().UnixNano()
			fmt.Fprint(os.Stderr, (nowTime-startTime)/1000000)
			// needsShutdown <- true
			// panic("bye")
		}()

		for {
			bs := make([]byte, 1000)
			_, err := ptmx.Read(bs)
			if err != nil {
				if err.Error() == "read /dev/ptmx: input/output error" {
					break
				} else {
					panic(err)
				}
			}
			for _, b := range bs {
				if b == '\x60' {
					return
				}
				stdout <- rune(b)
			}
		}
	})()

	return Shell{
		stdout: stdout,
		ptmx:   ptmx,
		cmd:    cmd,
	}
}

// Kill safely shuts down the shell, closing stdout
func (s *Shell) Kill() {
	close(s.stdout)

	err := s.ptmx.Close()
	if err != nil {
		log.Fatal("failed to close ptmx", err)
	}

	err = s.cmd.Process.Kill()
	if err != nil {
		log.Fatal("failed to kill term process", err)
	}
}

func (s *Shell) handleStdin(data string) {
	_, err := s.ptmx.Write([]byte(data))
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Shell) resize(w, h int) {
	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			err := pty.Setsize(s.ptmx, &pty.Winsize{
				Rows: uint16(h), Cols: uint16(w),
				X: 16 * uint16(w), Y: 16 * uint16(h),
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.
}
