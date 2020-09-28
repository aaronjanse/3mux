package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func attach(sessionInfo SessionInfo) {
	fdSocketPath := path.Join(sessionInfo.path, "fd.sock")
	fmt.Printf("Waiting for server to be online... (%s)\n", fdSocketPath)

	waitForFdSock(sessionInfo, fdSocketPath)

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}

	fmt.Print("\x1b[?1049h")
	fmt.Print("\x1b[?1006h")
	fmt.Print("\x1b[?1002h")

	restoreTerminal := func() {
		terminal.Restore(0, oldState)
		fmt.Print("\x1b[?1049l")
		fmt.Print("\x1b[?1006l")
		fmt.Print("\x1b[?1002l")
	}

	defer restoreTerminal()

	fmt.Print("\x1b[?1l")

	fdConn, err := net.Dial("unix", fdSocketPath)
	if err != nil {
		restoreTerminal()
		fmt.Println("Although the server socket exists, connection to it failed:", err)
		_, err = net.Dial("unix", path.Join(sessionInfo.path, "kill-server.sock"))
		if err != nil {
			fmt.Println("Killing the server failed.")
		}
		fmt.Println("See logs at", path.Join(sessionInfo.path, "logs-server.txt"))
		fmt.Println("To allow a new session to be created with this name, run:")
		fmt.Printf("$ rm -rf %s\n", sessionInfo.path)
		os.Exit(1)
	}
	fConn, err := fdConn.(*net.UnixConn).File()
	if err != nil {
		restoreTerminal()
		fmt.Println("After connection to the server socket, processing failed:", err)
		_, err = net.Dial("unix", path.Join(sessionInfo.path, "kill-server.sock"))
		if err != nil {
			fmt.Println("Killing the server failed.")
		}
		fmt.Println("See logs at", path.Join(sessionInfo.path, "logs-server.txt"))
		os.Exit(1)
	}

	// defer net.Dial("unix", path.Join(dir, "detach-server.sock"))
	rights := syscall.UnixRights(int(os.Stdin.Fd()), int(os.Stdout.Fd()))
	err = syscall.Sendmsg(int(fConn.Fd()), nil, rights, nil, 0)
	if err != nil {
		restoreTerminal()
		fmt.Println("Passing terminal control to the session server failed:", err)
		_, err = net.Dial("unix", path.Join(sessionInfo.path, "kill-server.sock"))
		if err != nil {
			fmt.Println("Killing the server failed.")
		}
		fmt.Println("See logs at", path.Join(sessionInfo.path, "logs-server.txt"))
		os.Exit(1)
	}
	fConn.Close()

	go func() {
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGWINCH)
			<-c
			updateSize(sessionInfo.path)
		}
	}()

	updateSize(sessionInfo.path)

	killClientPath := path.Join(sessionInfo.path, "kill-client.sock")
	os.Remove(killClientPath)
	detachSocket, err := net.Listen("unix", killClientPath)
	if err != nil {
		net.Dial("unix", path.Join(sessionInfo.path, "detach-server.sock"))
		restoreTerminal()
		fmt.Println("Client shutdown scenario planning failed:", err)
		_, err = net.Dial("unix", path.Join(sessionInfo.path, "kill-server.sock"))
		if err != nil {
			fmt.Println("Killing the server failed.")
		}
		fmt.Println("See logs at", path.Join(sessionInfo.path, "logs-server.txt"))
		os.Exit(1)
	}
	_, err = detachSocket.Accept()
	if err != nil {
		net.Dial("unix", path.Join(sessionInfo.path, "detach-server.sock"))
		restoreTerminal()
		fmt.Println("Waiting for client shutdown failed:", err)
		_, err = net.Dial("unix", path.Join(sessionInfo.path, "kill-server.sock"))
		if err != nil {
			fmt.Println("Killing the server failed.")
		}
		fmt.Println("See logs at", path.Join(sessionInfo.path, "logs-server.txt"))
		os.Exit(1)
	}
}

func waitForFdSock(sessionInfo SessionInfo, fdSocketPath string) {
	for i := 10; ; i-- {
		fdata, err := os.Stat(fdSocketPath)
		if err == nil && fdata != nil {
			break
		}
		time.Sleep(time.Millisecond * 50)

		if i == 0 {
			fmt.Println("Server failed to boot.")
			serverLogsPath := path.Join(sessionInfo.path, "logs-server.txt")
			if fdata, err := os.Stat(serverLogsPath); err != nil && fdata == nil {
				fmt.Println("The server never wrote logs.")
				err = os.RemoveAll(sessionInfo.path)
				if err != nil {
					fmt.Printf(
						"Failed to remove metadata directory `%s`: %s\n",
						sessionInfo.path, err.Error(),
					)
				}
			} else {
				fmt.Println("Server logs can be found at:", serverLogsPath)
				fmt.Println("To allow a new session to be created with this name, run:")
				fmt.Printf("$ rm -rf %s\n", sessionInfo.path)
			}
			os.Exit(1)
		}
	}
}

func updateSize(dir string) {
	w, h, _ := getTermSize()

	conn, err := net.Dial("unix", path.Join(dir, "resize.sock"))
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
