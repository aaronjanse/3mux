package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/wm"
	"github.com/npat-efault/poller"
)

func serve(sessionID string) {
	var clientExitSock = "booted.sock"

	dir := path.Join(threemuxDir, sessionID)
	os.MkdirAll(dir, 0755)
	stateBeforeInput := ""

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			diagnostics := fmt.Sprintf(
				"3mux has encountered a fatal error: %s\n"+
					"%s\n\n"+
					"Window manager state before error: `%s`\n\n"+
					"Please report this to https://github.com/aaronjanse/3mux/issues.",
				r.(error).Error(), string(debug.Stack()), stateBeforeInput,
			)

			log.Println(diagnostics)

			// tell the client to gracefully detach then warn the user
			conn, err := net.Dial("unix", path.Join(dir, clientExitSock))
			if err == nil {
				conn.Write([]byte(diagnostics))
			}

			timestamp := time.Now().UTC().Format(time.RFC3339)
			diagnostics = fmt.Sprintf("At %s...\n\n%s", timestamp, diagnostics)
			ioutil.WriteFile(path.Join(dir, "crash"), []byte(diagnostics), 0666)
		} else {
			// tell the client to gracefully detach
			net.Dial("unix", path.Join(dir, clientExitSock))

			os.RemoveAll(dir)
		}
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

	var parser *ecma48.Parser

	var stdinBuf *bufio.Reader
	var stdinPoller *poller.FD
	listenFd(sessionID, func(stdinFd, stdoutFd int) {
		stdinPoller, _ = poller.NewFD(stdinFd)
		stdinBuf = bufio.NewReader(stdinPoller)
		parser = ecma48.NewParser(true)
		go parser.Parse(stdinBuf, stdin)
		renderer.UpdateOut(stdoutFd)
	})

	listenResize(sessionID, func(width, height int) {
		renderer.Resize(width, height)
		u.SetRenderRect(0, 0, width, height)
	})

	go func() {
		defer recover()
		detachSocket, err := net.Listen("unix", path.Join(dir, "detach-server.sock"))
		if err != nil {
			panic(err)
		}
		for {
			detachSocket.Accept()

			renderer.UpdateOut(-1)
			parser.Shutdown <- nil
			stdinPoller.Close()
		}
	}()

	go func() {
		detachSocket, err := net.Listen("unix", path.Join(dir, "kill-server.sock"))
		if err != nil {
			panic(err)
		}
		detachSocket.Accept()

		shutdown <- nil
	}()

	go net.Dial("unix", path.Join(dir, "booted.sock"))
	clientExitSock = "shutdown.sock"

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
	sockPath := path.Join(threemuxDir, sessionID, "resize.sock")
	socket, err := net.Listen("unix", sockPath)
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
	sockPath := path.Join(threemuxDir, sessionID, "fd.sock")
	socket, err := net.Listen("unix", sockPath)
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
