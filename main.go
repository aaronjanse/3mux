package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aaronjanse/3mux/ecma48"
	"github.com/npat-efault/poller"
	"github.com/sevlyar/go-daemon"
	"golang.org/x/crypto/ssh/terminal"
)

type fdReader int

func (fd fdReader) Read(p []byte) (n int, err error) {
	return syscall.Read(int(fd), p)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var writeLogs = flag.Bool("log", false, "write logs to ./logs.txt")

var cmdNew = flag.NewFlagSet("new", flag.ExitOnError)
var cmdNewDetach = cmdNew.Bool("detach", false, "start session without attaching")

var cmdHelp = flag.Bool("help", false, "show help")
var cmdHelpShort = flag.Bool("h", false, "show help")

const dir = "/tmp/3mux/"

func main() {
	flag.Parse()

	// setup logging
	if *writeLogs {
		f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// setup cpu profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var cmd string
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}

	if *cmdHelp || *cmdHelpShort {
		showHelp()
		return
	}

	parentSessionID := os.Getenv("THREEMUX")

	switch cmd {
	case "help":
		showHelp()
	case "detach":
		if parentSessionID == "" {
			fmt.Println("Must be within session to detach")
			return
		}
		detach(parentSessionID)
	case "_serve-id":
		sessionID := os.Args[2]
		serve(sessionID)
	case "new":
		if parentSessionID != "" {
			refuseNesting()
		}
		sessionName := os.Args[2]
		sessionID := launchServer(sessionName)
		attach(sessionID)
	case "ls":
		os.MkdirAll("/tmp/3mux", 0755)
		children, err := ioutil.ReadDir("/tmp/3mux/")
		if err != nil {
			panic(err)
		}
		fmt.Println("Sessions:")
		for _, child := range children {
			dir := fmt.Sprintf("/tmp/3mux/%s/", child.Name())
			thisName, _ := ioutil.ReadFile(dir + "name")
			fmt.Printf("- %s\n", string(thisName))
		}
	case "attach":
		if parentSessionID != "" {
			refuseNesting()
		}

		sessionName := os.Args[2]
		sessionID, err := findSession(sessionName)
		if err != nil {
			fmt.Printf("Could not find session `%s`\n", sessionName)
		}
		attach(sessionID)
	default:
		if parentSessionID == "" {
			sessionName, isNew := defaultPrompt()
			if sessionName == "" {
				return
			}

			var sessionID string
			if isNew {
				sessionID = launchServer(sessionName)
			} else {
				var err error
				sessionID, err = findSession(sessionName)
				if err != nil {
					fmt.Printf("Could not find session `%s`\n", sessionName)
				}
			}

			attach(sessionID)
		} else {
			dir := fmt.Sprintf("/tmp/3mux/%s/", parentSessionID)
			thisName, _ := ioutil.ReadFile(dir + "name")
			fmt.Printf("You're in session `%s`\n", string(thisName))
			var choice string
			for choice != "y" && choice != "n" {
				fmt.Print("Would you like to detach? [Y/n] ")
				fmt.Scanf("%s", &choice)
				choice = strings.ToLower(choice)
				if choice == "" {
					choice = "y"
				}
			}
			if choice == "y" {
				detach(parentSessionID)
			}
		}
	}
}

func detach(parentSessionID string) {
	net.Dial("unix", fmt.Sprintf("/tmp/3mux/%s/detach-client.sock", parentSessionID))
}

func findSession(sessionName string) (sessionID string, err error) {
	os.MkdirAll("/tmp/3mux", 0666)
	children, err := ioutil.ReadDir("/tmp/3mux/")
	if err != nil {
		panic(err)
	}

	for _, child := range children {
		id := child.Name()
		dir := fmt.Sprintf("/tmp/3mux/%s/", id)
		nameRaw, _ := ioutil.ReadFile(dir + "name")
		nameCleaned := strings.TrimSpace(string(nameRaw))

		if nameCleaned == sessionName {
			return id, nil
		}
	}

	return "", fmt.Errorf("Could not find session: `%s`", sessionName)
}

func launchServer(sessionName string) (sessionID string) {
	os.MkdirAll("/tmp/3mux", 0777)

	id := 0
	for {
		path := fmt.Sprintf("/tmp/3mux/%d/", id)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		id++
	}
	sessionID = strconv.Itoa(id)

	dir := fmt.Sprintf("/tmp/3mux/%s/", sessionID)
	os.MkdirAll(dir, 0755)                                  // FIXME perms
	ioutil.WriteFile(dir+"name", []byte(sessionName), 0777) // FIXME perms

	readySocket, err := net.Listen("unix", dir+"ready.sock")
	if err != nil {
		panic(err)
	}

	args := []string{"3mux", "_serve-id", sessionID}
	if *writeLogs {
		args = append(args, "--log")
	}

	context := &daemon.Context{
		// PidFileName: "sample.pid",
		Args: args,
	}
	_, err = context.Reborn()
	if err != nil {
		panic(err)
	}

	readySocket.Accept()
	os.Remove(dir + "ready.sock")

	return sessionID
}

func refuseNesting() {
	fmt.Println("Refusing to run 3mux inside itself.")
	fmt.Println("If you want to do so anyway, `unset THREEMUX`.")
}

func defaultPrompt() (sessionName string, isNew bool) {
	os.MkdirAll("/tmp/3mux", 0755)
	children, err := ioutil.ReadDir("/tmp/3mux/")
	if err != nil {
		panic(err)
	}

	optionIdx := 0
	options := []string{}
	idxToDir := map[int]string{}

	for idx, child := range children {
		dir := fmt.Sprintf("/tmp/3mux/%s/", child.Name())
		thisName, _ := ioutil.ReadFile(dir + "name")

		options = append(options, strings.TrimSpace(string(thisName)))

		idxToDir[idx] = dir
	}

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)

	stdin := make(chan ecma48.Output, 64)
	defer close(stdin)

	parser := ecma48.NewParser(true)
	pollerIn, _ := poller.NewFD(int(os.Stdin.Fd()))

	go parser.Parse(bufio.NewReader(pollerIn), stdin)
	defer func() {
		parser.Shutdown <- nil
		// don't close because we need stdin to be available for the daemon later
		pollerIn.SetReadDeadline(time.Now())
	}()

	if len(options) == 0 {
		isNew = true
		sessionName = promptNewSessionName(options, stdin)
		return
	} else {
		fmt.Print("Attach to an existing session or create a new one:\r\n")

		// fmt.Print("\x1b[?25l") // hide cursor

		fmt.Print("\x1b[1;36m> " + options[0] + "\x1b[39;2m\r\n")

		for _, option := range options[1:] {
			fmt.Print("  " + option + "\x1b[39m\r\n")
		}

		fmt.Print("+ create new session\r")
		fmt.Printf("\x1b[%dA", len(options))

		clearOptions := func() {
			fmt.Printf("\r\x1b[%dB", len(options)-optionIdx)
			for i := 0; i < len(options)+1; i++ {
				fmt.Print("\x1b[K\x1b[A")
			}
			fmt.Print("\x1b[K")
		}

		for {
			next := <-stdin
			if len(next.Raw) == 1 {
				switch next.Raw[0] {
				case 13:
					clearOptions()
					if optionIdx == len(options) {
						isNew = true
						sessionName = promptNewSessionName(options, stdin)
						return
					} else if optionIdx < len(options) {
						sessionName = options[optionIdx]
						return
					}
				case 3:
					clearOptions()
					return
				}
			}
			switch x := next.Parsed.(type) {
			case ecma48.CursorMovement:
				switch x.Direction {
				case ecma48.Up:
					if optionIdx > 0 {
						if optionIdx == len(options) {
							fmt.Print("\r+ create new session\r")
						} else {
							fmt.Print("\r  " + options[optionIdx] + "\r")
						}
						optionIdx--
						fmt.Print("\x1b[1A")
						fmt.Print("\x1b[22;1;36m> " + options[optionIdx] + "\x1b[39;2m\r")
					}
				case ecma48.Down:
					if optionIdx < len(options) {
						fmt.Print("\r  " + options[optionIdx] + "\r\n")
						optionIdx++
						if optionIdx == len(options) {
							fmt.Print("\x1b[22;1;36m+ create new session\x1b[39;2m\r")
						} else {
							fmt.Print("\x1b[22;1;36m> " + options[optionIdx] + "\x1b[39;2m\r")
						}
					}
				}
			}
		}
	}
}

func promptNewSessionName(existing []string, stdin chan ecma48.Output) string {
	problem := ""
	for {
		fmt.Print("\x1b[22mName of new session:\r\n")
		fmt.Print("\x1b[22;36m? \x1b[m")
		fmt.Print(problem)
		fmt.Print("\r\x1b[2C")

		name := surveyName(stdin)
		if name == "" {
			return ""
		}

		for _, existingName := range existing {
			if name == existingName {
				problem = "session names must be unique"
				continue
			}
		}

		return name
	}
}

func surveyName(stdin chan ecma48.Output) string {
	defer fmt.Print("\r\x1b[K\x1b[A\x1b[K")
	idx := 0
	out := ""
	for {
		next := <-stdin
		if len(next.Raw) == 1 {
			switch next.Raw[0] {
			case 0:
				continue
			case 3:
				return ""
			case 13:
				if out != "" {
					return strings.TrimSpace(out)
				}
			}
		}
		switch x := next.Parsed.(type) {
		case ecma48.Char:
			fmt.Print(string(x.Rune))
			fmt.Print(string(out[idx:]))
			le := len(out[idx:])
			if le > 0 {
				fmt.Printf("\x1b[%dD", le)
			}
			out = out[:idx] + string(x.Rune) + out[idx:]
			// fmt.Print("[", (out, idx, "]")
			idx++
		case ecma48.Backspace:
			if idx > 0 {
				out = out[:idx-1] + out[idx:]
				idx--
				fmt.Print("\b\x1b[K" + string(out[idx:]))
				le := len(out[idx:])
				if le > 0 {
					fmt.Printf("\x1b[%dD", le)
				}
				// }
			}
		case ecma48.CursorMovement:
			switch x.Direction {
			case ecma48.Left:
				if idx > 0 {
					fmt.Print("\x1b[1D")
					idx--
				}
			case ecma48.Right:
				if idx < len(out) {
					fmt.Print("\x1b[1C")
					idx++
				}
			}
		}
	}
}
