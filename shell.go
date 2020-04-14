package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	runtimeDebug "runtime/debug"
	"sync/atomic"
	"syscall"

	"github.com/kr/pty"
)

// Shell manages spawning, killing, and sending data to/from a shell subprocess (e.g. bash, sh, zsh)
type Shell struct {
	stdout      chan<- rune
	ptmx        *os.File
	cmd         *exec.Cmd
	byteCounter uint64
}

func newShell(stdout chan<- rune) Shell {
	cmd := exec.Command(os.Getenv("SHELL"))

	cmd.Env = append(os.Environ(),
		"THREEMUX=1",
	)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		fatalShutdownNow(err.Error())
	}

	shell := Shell{
		stdout: stdout,
		ptmx:   ptmx,
		cmd:    cmd,
	}

	// feed ptmx output to stdout channel
	go (func() {
		defer func() {
			if r := recover(); r != nil {
				// FIXME: bad practice
				if r.(error).Error() != "send on closed channel" &&
					r.(error).Error() != "read /dev/ptmx: file already closed" {
					fatalShutdownNow("shell.go\n" + r.(error).Error())
				}
			}
		}()

		for {
			scanner := bufio.NewReader(ptmx)
			for {
				if r, _, err := scanner.ReadRune(); err != nil {
					if err.Error() == "read /dev/ptmx: input/output error" {
						break // ^D
					} else if err == io.EOF {
						break
					} else {
						panic(err)
					}
				} else {
					atomic.AddUint64(&shell.byteCounter, 1)
					stdout <- r
				}
			}
		}
	})()

	return shell
}

// Kill safely shuts down the shell, closing stdout
func (s *Shell) Kill() {
	close(s.stdout)

	err := s.ptmx.Close()
	if err != nil {
		fatalShutdownNow("failed to close ptmx; " + err.Error())
	}

	err = s.cmd.Process.Kill()
	if err != nil { // FIXME
		log.Println("failed to kill term process", err)
	}
}

func (s *Shell) handleStdin(data string) {
	_, err := s.ptmx.Write([]byte(data))
	if err != nil {
		fatalShutdownNow(err.Error())
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
				log.Printf("Error during: shell.go:resize(%d, %d)", w, h)
				log.Println("Tiling state:", root.serialize())
				log.Println(string(runtimeDebug.Stack()))
				log.Println()
				log.Println("Please submit a bug report with this stack trace to https://github.com/aaronjanse/3mux/issues")
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.
}
