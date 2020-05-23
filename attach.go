package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func attach(sessionID string) {
	dir := fmt.Sprintf("/tmp/3mux/%s/", sessionID)

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)

	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")
	fmt.Print("\x1b[?1006h")
	defer fmt.Print("\x1b[?1006l")
	fmt.Print("\x1b[?1002h")
	defer fmt.Print("\x1b[?1002l")

	fmt.Print("\x1b[?1l")

	fdConn, err := net.Dial("unix", dir+"fd.sock")
	if err != nil {
		panic(err)
	}
	fConn, err := fdConn.(*net.UnixConn).File()
	if err != nil {
		panic(err)
	}

	sendFds := func(in, out *os.File) {
		rights := syscall.UnixRights(int(in.Fd()), int(out.Fd()))
		err = syscall.Sendmsg(int(fConn.Fd()), nil, rights, nil, 0)
		if err != nil {
			panic(err)
		}
		fConn.Close()
	}

	sendFds(os.Stdin, os.Stdout)
	// os.Stdin.Write([]byte{1})

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			updateSize(dir)
		}
	}()

	updateSize(dir)

	os.Remove(dir + "detach-client.sock")
	detachSocket, err := net.Listen("unix", dir+"detach-client.sock")
	if err != nil {
		panic(err)
	}
	detachSocket.Accept()

	net.Dial("unix", dir+"detach-server.sock")
}

func updateSize(dir string) {
	w, h, _ := getTermSize()

	conn, err := net.Dial("unix", dir+"resize.sock")
	if err != nil {
		panic(err)
	}

	conn.Write([]byte{
		byte(w >> 8), byte(w % 256),
		byte(h >> 8), byte(h % 256),
	})

	conn.Close()
}

func getTermSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	outStr := strings.TrimSpace(string(out))
	parts := strings.Split(outStr, " ")

	h, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	w, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	wInt := int(int64(w))
	hInt := int(int64(h))
	return wInt, hInt, nil
}
