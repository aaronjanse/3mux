package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
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

var threemuxDir string

func main() {
	tmp := os.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}

	threemuxDir = path.Join(tmp, "3mux")

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
			return
		}
		sessionName := os.Args[2]
		sessionID := launchServer(sessionName)
		attach(sessionID)
	case "ls", "ps":
		os.MkdirAll(threemuxDir, 0755)
		children, err := ioutil.ReadDir(threemuxDir)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sessions:")
		for _, child := range children {
			namePath := path.Join(threemuxDir, child.Name(), "name")
			thisName, _ := ioutil.ReadFile(namePath)
			fmt.Printf("- %s\n", string(thisName))
		}
	case "attach":
		if parentSessionID != "" {
			refuseNesting()
			return
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
			namePath := path.Join(threemuxDir, parentSessionID, "name")
			thisName, _ := ioutil.ReadFile(namePath)
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
	net.Dial("unix", path.Join(threemuxDir, parentSessionID, "shutdown.sock"))
}

func findSession(sessionName string) (sessionID string, err error) {
	os.MkdirAll(threemuxDir, 0666)
	children, err := ioutil.ReadDir(threemuxDir)
	if err != nil {
		panic(err)
	}

	for _, child := range children {
		id := child.Name()
		namePath := path.Join(threemuxDir, id, "name")
		nameRaw, _ := ioutil.ReadFile(namePath)
		nameCleaned := strings.TrimSpace(string(nameRaw))

		if nameCleaned == sessionName {
			return id, nil
		}
	}

	return "", fmt.Errorf("Could not find session: `%s`", sessionName)
}

func launchServer(sessionName string) (sessionID string) {
	os.MkdirAll(threemuxDir, 0777)

	id := 0
	for {
		dir := path.Join(threemuxDir, strconv.Itoa(id))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			break
		}
		id++
	}
	sessionID = strconv.Itoa(id)

	dir := path.Join(threemuxDir, sessionID)
	os.MkdirAll(dir, 0755)                                              // FIXME perms
	ioutil.WriteFile(path.Join(dir, "name"), []byte(sessionName), 0777) // FIXME perms

	readySocket, err := net.Listen("unix", path.Join(dir, "booted.sock"))
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

	conn, _ := readySocket.Accept()
	logs, _ := ioutil.ReadAll(conn)
	if len(logs) > 0 {
		fmt.Print(string(logs))
		os.Exit(1)
	}

	os.Remove(path.Join(dir, "booted.sock"))

	return sessionID
}

func refuseNesting() {
	fmt.Println("Refusing to run 3mux inside itself.")
	fmt.Println("If you want to do so anyway, `unset THREEMUX`.")
}

type SessionEntry struct {
	name    string
	crashed bool
}

func defaultPrompt() (sessionName string, isNew bool) {
	os.MkdirAll(threemuxDir, 0755)
	children, err := ioutil.ReadDir(threemuxDir)
	if err != nil {
		panic(err)
	}

	optionIdx := 0
	options := []SessionEntry{}
	idxToDir := map[int]string{}

	for idx, child := range children {
		dir := path.Join(threemuxDir, child.Name())
		thisName, _ := ioutil.ReadFile(path.Join(dir, "name"))

		// if we can successfully `stat` the crash log, then the server crashed
		_, statCrashLogErr := os.Stat(path.Join(dir, "crash"))
		crashed := statCrashLogErr == nil

		options = append(options, SessionEntry{
			name:    strings.TrimSpace(string(thisName)),
			crashed: crashed,
		})

		idxToDir[idx] = dir
	}

	formatSessionEntry := func(entry SessionEntry, selected bool) string {
		var format string
		if entry.crashed {
			if selected {
				format = "\x1b[;1;91m? %s (crashed)\x1b[39;2m"
			} else {
				format = "\x1b[;93m  %s (crashed)\x1b[39;2m"
			}
		} else {
			if selected {
				format = "\x1b[;1;36m> %s\x1b[39;2m"
			} else {
				format = "\x1b[m  %s\x1b[39;2m"
			}
		}
		return fmt.Sprintf(format, entry.name)
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

		fmt.Print(formatSessionEntry(options[0], true) + "\r\n")

		for _, option := range options[1:] {
			fmt.Print(formatSessionEntry(option, false) + "\r\n")
		}

		fmt.Print("+ create new session\r")
		fmt.Printf("\x1b[%dA", len(options))
		fmt.Printf("\x1b[m")

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
						sessionName = options[optionIdx].name
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
							fmt.Print(formatSessionEntry(options[optionIdx], false) + "\r")
						}
						optionIdx--
						fmt.Print("\x1b[1A")
						fmt.Print(formatSessionEntry(options[optionIdx], true) + "\r")
					}
				case ecma48.Down:
					if optionIdx < len(options) {
						fmt.Print(formatSessionEntry(options[optionIdx], false) + "\r\n")
						optionIdx++
						if optionIdx == len(options) {
							fmt.Print("\x1b[22;1;36m+ create new session\x1b[39;2m\r")
						} else {
							fmt.Print(formatSessionEntry(options[optionIdx], true) + "\r")
						}
					}
				}
			}
		}
	}
}

func promptNewSessionName(existing []SessionEntry, stdin chan ecma48.Output) string {
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
			if name == existingName.name {
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
