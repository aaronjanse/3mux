package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func attach(sessionInfo *SessionInfo) error {
	fmt.Printf("Waiting for server to be online... (%s)\n", sessionInfo.fdPath)

	err := waitForFdSock(sessionInfo)
	if err != nil {
		return err
	}

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return errors.New("failed to enable terminal raw mode")
	}
	defer terminal.Restore(0, oldState)

	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")
	fmt.Print("\x1b[?1006h")
	defer fmt.Print("\x1b[?1006l")
	fmt.Print("\x1b[?1002h")
	defer fmt.Print("\x1b[?1002l")

	fmt.Print("\x1b[?1l")

	fdConn, err := net.Dial("unix", sessionInfo.fdPath)
	if err != nil {
		return fmt.Errorf("Although the server socket exists, connection to it failed: %s", err)
	}
	fConn, err := fdConn.(*net.UnixConn).File()
	if err != nil {
		return fmt.Errorf("After connection to the server socket, processing failed: %s", err)
	}

	rights := syscall.UnixRights(int(os.Stdin.Fd()), int(os.Stdout.Fd()))
	err = syscall.Sendmsg(int(fConn.Fd()), nil, rights, nil, 0)
	if err != nil {
		return fmt.Errorf("Passing terminal control to the session server failed: %s", err)
	}
	fConn.Close()

	defer net.Dial("unix", sessionInfo.detachPath)

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			updateSize(sessionInfo)
		}
	}()

	updateSize(sessionInfo)

	os.Remove(sessionInfo.killClientPath)
	detachSocket, err := net.Listen("unix", sessionInfo.killClientPath)
	if err != nil {
		return fmt.Errorf("Client shutdown scenario planning failed: %s", err)
	}
	_, err = detachSocket.Accept()
	if err != nil {
		return fmt.Errorf("Waiting for client shutdown failed: %s", err)
	}

	return nil
}

func waitForFdSock(sessionInfo *SessionInfo) error {
	for i := 10; ; i-- {
		fdata, err := os.Stat(sessionInfo.fdPath)
		if err == nil && fdata != nil {
			break
		}
		time.Sleep(time.Millisecond * 50)

		if i == 0 {
			return fmt.Errorf("Server failed to boot in 0.5s")
		}
	}
	return nil
}

func updateSize(sessionInfo *SessionInfo) {
	w, h, _ := getTermSize()

	conn, err := net.Dial("unix", sessionInfo.resizePath)
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
