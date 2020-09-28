package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"syscall"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/aaronjanse/3mux/pane"
	"github.com/aaronjanse/3mux/render"
	"github.com/aaronjanse/3mux/wm"
	"github.com/npat-efault/poller"
)

func serve(sessionInfo SessionInfo) error {
	log.Println("Booting...")

	config, err := loadOrGenerateConfig()
	if err != nil {
		return fmt.Errorf("Failed to load or generate config: %s", err)
	}

	renderer := render.NewRenderer(-1)
	renderer.Resize(20, 20)
	go renderer.ListenToQueue()

	shutdown := make(chan error)

	newPane := func(renderer ecma48.Renderer) wm.Node {
		return pane.NewPane(renderer, true, sessionInfo.uuid)
	}

	u := wm.NewUniverse(renderer,
		config.generalSettings.EnableHelpBar,
		config.generalSettings.EnableStatusBar,
		func(err error) {
			go func() {
				if err != nil {
					shutdown <- fmt.Errorf("%s\n%s", err, debug.Stack())
				} else {
					shutdown <- nil
				}
			}()
		}, wm.Rect{X: 0, Y: 0, W: 50, H: 20}, newPane)
	defer u.Kill()

	stdin := make(chan ecma48.Output, 64)
	defer close(stdin)

	var parser *ecma48.Parser

	var stdinBuf *bufio.Reader
	var stdinPoller *poller.FD
	listenFd(sessionInfo, func(stdinFd, stdoutFd int) {
		stdinPoller, _ = poller.NewFD(stdinFd)
		stdinBuf = bufio.NewReader(stdinPoller)
		parser = ecma48.NewParser(true)
		go parser.Parse(stdinBuf, stdin)
		renderer.UpdateOut(stdoutFd)
	})

	defer net.Dial("unix", sessionInfo.killClientPath)

	listenResize(sessionInfo, func(width, height int) {
		renderer.Resize(width, height)
		u.SetRenderRect(0, 0, width, height)
	})

	go func() {
		detachSocket, err := net.Listen("unix", sessionInfo.detachPath)
		if err != nil {
			log.Println("Detach error:", err)
			panic(err)
		}
		for {
			_, err = detachSocket.Accept()
			if err != nil {
				log.Println("Detach accept error:", err)
				panic(err)
			}

			log.Println("Detaching...")

			renderer.UpdateOut(-1)
			parser.Shutdown <- nil
			stdinPoller.Close()

			log.Println("Detaching... DONE")
			net.Dial("unix", sessionInfo.killClientPath)
		}
	}()

	go func() {
		detachSocket, err := net.Listen("unix", sessionInfo.killServerPath)
		if err != nil {
			panic(err)
		}
		detachSocket.Accept()

		shutdown <- nil
	}()

	for {
		select {
		case next := <-stdin:
			human := humanify(next)
			log.Println("Keypress:", human)

			if human == "Ctrl+Q" {
				return nil
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
				return err
			}
			return nil
		}
	}
}

func listenResize(sessionInfo SessionInfo, callback func(width, height int)) {
	socket, err := net.Listen("unix", sessionInfo.resizePath)
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

func listenFd(sessionInfo SessionInfo, callback func(stdinFd, stdoutFd int)) {
	socket, err := net.Listen("unix", sessionInfo.fdPath)
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
	if err != nil {
		panic(err)
	}
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
