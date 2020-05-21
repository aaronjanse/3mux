package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/wm"
)

func serve(sessionID string) {
	fmt.Println("Serving!")
	dir := fmt.Sprintf("/tmp/3mux/%s/", sessionID)
	os.MkdirAll(dir, 0755)
	stateBeforeInput := ""

	defer func() {
		fmt.Println("Dying!")
		if r := recover(); r != nil {
			fmt.Println(r)
			diagnostics := fmt.Sprintf(
				"3mux has encountered a fatal error: %s\n"+
					"%s\n\n"+
					"Window manager state before error: `%s`\n\n"+
					"Please report this to https://github.com/aaronjanse/3mux/issues.",
				r.(error).Error(), string(debug.Stack()), stateBeforeInput,
			)

			// tell the client to gracefully detach then warn the user
			conn, err := net.Dial("unix", dir+"fatal.sock")
			if err == nil {
				conn.Write([]byte(diagnostics))
			} else {
				timestamp := time.Now().UTC().Format(time.RFC3339)
				diagnostics = fmt.Sprintf("At %s...\n\n%s", timestamp, diagnostics)
				ioutil.WriteFile(dir+"crash", []byte(diagnostics), 0666)
			}
		} else {
			// tell the client to gracefully detach
			net.Dial("unix", dir+"detach-client.sock")
		}

		os.RemoveAll(dir)
	}()

	config := loadOrGenerateConfig()

	renderer := render.NewRenderer(-1)
	renderer.Resize(20, 20)
	go renderer.ListenToQueue()

	shutdown := make(chan error)

	newPane := func(renderer ecma48.Renderer) wm.Node {
		return pane.NewPane(renderer, true, sessionID)
	}

	u := wm.NewUniverse(renderer,
		config.generalSettings.EnableHelpBar,
		config.generalSettings.EnableStatusBar,
		func(err error) {
			go func() {
				if err != nil {
					shutdown <- fmt.Errorf("%s\n%s", err.Error(), debug.Stack())
				} else {
					shutdown <- nil
				}
			}()
		}, wm.Rect{X: 0, Y: 0, W: 20, H: 20}, newPane)
	defer u.Kill()

	stdin := make(chan ecma48.Output, 64)
	defer close(stdin)

	fmt.Print("Creating init parser... ")
	var parser *ecma48.Parser
	fmt.Println("DONE")

	var oldStdinFd int
	listenFd(sessionID, func(stdinFd, stdoutFd int) {
		oldStdinFd = stdinFd
		fmt.Println("STDIN:  ", stdinFd)
		fmt.Println("STDOUT: ", stdoutFd)
		fmt.Print("Updating fds... ")
		stdinBuf := bufio.NewReader(fdReader(stdinFd))
		parser = ecma48.NewParser(true)
		go parser.Parse(stdinBuf, stdin)
		renderer.UpdateOut(stdoutFd)
		fmt.Println("DONE")
	})

	listenResize(sessionID, func(width, height int) {
		fmt.Print("Updating size... ")
		renderer.Resize(width, height)
		u.SetRenderRect(0, 0, width, height)
		fmt.Println("DONE")
	})

	go func() {
		detachSocket, err := net.Listen("unix", dir+"detach-server.sock")
		if err != nil {
			panic(err)
		}
		for {
			detachSocket.Accept()

			fmt.Print("Detaching fds... ")
			renderer.UpdateOut(-1)
			// parser.Shutdown <- nil
			syscall.Close(oldStdinFd)
			fmt.Println("DONE")
		}
	}()

	go func() {
		detachSocket, err := net.Listen("unix", dir+"kill-server.sock")
		if err != nil {
			panic(err)
		}
		detachSocket.Accept()

		shutdown <- nil
	}()

	go net.Dial("unix", dir+"ready.sock")

	for {
		select {
		case next := <-stdin:
			stateBeforeInput = u.Serialize()

			human := humanify(next)

			if human == "Ctrl+Q" {
				return
			}

			if seiveMouseEvents(u, human, next) {
				break
			}
			if seiveConfigEvents(config, u, human) {
				break
			}

			// if we didn't find anything special, just pass the raw data to
			// the selected terminal
			u.HandleStdin(next)
		case err := <-shutdown:
			if err != nil {
				panic(err)
			} else {
				return
			}
		}
	}
}

func listenResize(sessionID string, callback func(width, height int)) {
	path := fmt.Sprintf("/tmp/3mux/%s/resize.sock", sessionID)
	socket, err := net.Listen("unix", path)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := socket.Accept()
			if err != nil {
				panic(err)
			}

			b := make([]byte, 4)
			_, err = conn.Read(b)
			if err != nil {
				panic(err)
			}

			width := (int(b[0]) << 8) + int(b[1])
			height := (int(b[2]) << 8) + int(b[3])

			callback(width, height)

			conn.Close()
		}
	}()
}

func listenFd(sessionID string, callback func(stdinFd, stdoutFd int)) {
	path := fmt.Sprintf("/tmp/3mux/%s/fd.sock", sessionID)
	socket, err := net.Listen("unix", path)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err := socket.Accept()
			if err != nil {
				panic(err)
			}
			callback(parseFdConn(conn.(*net.UnixConn)))
			conn.Close()
		}
	}()
}

func parseFdConn(conn *net.UnixConn) (stdinFd, stdoutFd int) {
	fConn, err := conn.File()
	if err != nil {
		panic(err)
	}

	numFds := 2
	buf := make([]byte, syscall.CmsgSpace(numFds*4))
	_, _, _, _, err = syscall.Recvmsg(int(fConn.Fd()), nil, buf, 0)
	if err != nil {
		panic(err)
	}
	msgs, err := syscall.ParseSocketControlMessage(buf)
	fds := []int{}
	for _, msg := range msgs {
		newFds, err := syscall.ParseUnixRights(&msg)
		if err != nil {
			panic(err)
		}
		fds = append(fds, newFds...)
	}

	stdinFd = fds[0]
	stdoutFd = fds[1]
	return
}
