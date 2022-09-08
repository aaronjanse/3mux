package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path"
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

var threemuxDir string

const BUG_REPORT_URL = "https://github.com/aaronjanse/3mux/issues"

func init() {
	tmp := os.Getenv("TMPDIR")
	if tmp == "" {
		tmp = "/tmp"
	}

	threemuxDir = path.Join(tmp, "3mux")
	os.MkdirAll(threemuxDir, 0700)
}

func main() {
	flag.Parse()

	parentSessionID := os.Getenv("THREEMUX")

	if len(os.Args) < 2 {
		if parentSessionID != "" {
			namePath := path.Join(threemuxDir, parentSessionID, "name")
			thisName, _ := ioutil.ReadFile(namePath)
			fmt.Printf("You're in session `%s`\n", string(thisName))
			var choice string
			for choice != "y" && choice != "n" {
				fmt.Print("Would you like to detach? [Y/n] ")
				fmt.Scanf("%s", &choice)
				choice = strings.TrimSpace(strings.ToLower(choice))
				if choice == "" {
					choice = "y"
				}
			}
			if choice == "y" {
				detach(parentSessionID)
				os.Exit(0)
			}
			os.Exit(1)
		}

		name, isNew := defaultPrompt()
		if name == "" { // only happens upon Ctrl-C
			os.Exit(1)
		}
		if isNew {
			os.Args = []string{os.Args[0], "new", name}
		} else {
			os.Args = []string{os.Args[0], "attach", name}
		}
	}

	switch os.Args[1] {
	case "new-server-internal-only":
		if len(os.Args) < 3 {
			fmt.Println("Missing `name` argument.")
			os.Exit(1)
		}
		sessionName := os.Args[2]

		daemonContext := &daemon.Context{
			Args: []string{
				os.Args[0],
				"new-server-internal-only",
				sessionName,
			},
		}
		defer daemonContext.Release()

		sessionInfo, found, err := findSession(sessionName)
		if !found || err != nil {
			fmt.Println("We couldn't find the session we're supposed to serve, so we couldn't even setup proper logging. Exiting.")
			os.Exit(1)
		}

		logsPath := path.Join(sessionInfo.path, "logs-server.txt")
		f, err := os.OpenFile(logsPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Printf("While initializing server logging, error opening file: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)

		child, err := daemonContext.Reborn()
		if err != nil {
			log.Println("Error occured while serving session daemon:", err)
			os.Exit(1)
		}
		if child != nil {
			log.Println("Encountered 'unreachable' state with server-client mismatch. Args:", os.Args)
			os.Exit(1)
		}

		err = serve(sessionInfo)
		if err == nil {
			log.Println("Exiting cleanly...")
			err := os.RemoveAll(sessionInfo.path)
			if err != nil {
				log.Printf("Failed to remove metadata directory `%s`: %s\n", sessionInfo.path, err)
				log.Printf("It can be removed manually with `rm -rf %s`, although the above error "+
					"message is likely relevant.", sessionInfo.path)
				os.Exit(1)
			}
		} else {
			log.Println(err)
			os.Exit(1)
		}
	case "new":
		if parentSessionID != "" {
			refuseNesting()
		}
		if len(os.Args) != 3 {
			fmt.Println("Usage: 3mux new <name>")
			os.Exit(1)
		}
		sessionName := os.Args[2]

		_, found, err := findSession(sessionName)
		if err != nil {
			fmt.Println("Error while querying sessions:", err)
			os.Exit(1)
		}
		if found {
			fmt.Println("Session with this name already exists:", sessionName)
			os.Exit(1)
		}

		initializeSession(sessionName)

		daemonContext := &daemon.Context{
			Args: []string{
				os.Args[0],
				"new-server-internal-only",
				sessionName,
			},
		}
		child, err := daemonContext.Reborn()
		if err != nil {
			fmt.Println("Error occured while spawning session daemon:", err)
			os.Exit(1)
		}
		if child == nil {
			fmt.Println("Encountered 'unreachable' state with server-client mismatch. Args:", os.Args)
			os.Exit(1)
		}

		os.Args[1] = "attach"
		fallthrough
	case "attach":
		if parentSessionID != "" {
			refuseNesting()
		}
		if len(os.Args) != 3 {
			fmt.Println("Usage: 3mux attach <name>")
			os.Exit(1)
		}
		sessionName := os.Args[2]
		sessionInfo, found, err := findSession(sessionName)
		if err != nil {
			fmt.Println("Error while querying sessions:", err)
			os.Exit(1)
		}
		if !found {
			fmt.Println("Failed to find session with name:", sessionName)
			os.Exit(1)
		}
		err = attach(sessionInfo)
		if err != nil {
			fmt.Println(err)
			fmt.Println("See server-side logs at", path.Join(sessionInfo.path, "logs-server.txt"))
			fmt.Printf("To manually kill this session, run `3mux kill %s`\n", sessionName)
			os.Exit(1)
		}
	case "kill":
		if len(os.Args) != 3 {
			fmt.Println("Usage: 3mux kill <name>")
			os.Exit(1)
		}
		sessionName := os.Args[2]
		sessionInfo, found, err := findSession(sessionName)
		if err != nil {
			fmt.Println("Error while querying sessions:", err)
			os.Exit(1)
		}
		if !found {
			fmt.Println("Failed to find session with name:", sessionName)
			os.Exit(1)
		}

		_, err = net.Dial("unix", sessionInfo.killServerPath)
		if err != nil {
			fmt.Println("Killing the server failed. Maybe one isn't running?")
			os.Exit(1)
		}

		err = os.RemoveAll(sessionInfo.path)
		if err != nil {
			fmt.Printf("Failed to remove metadata directory `%s`: %s\n\n", sessionInfo.path, err)
			fmt.Println("To create a new session with this name:")
			fmt.Println("1. Ensure there are no unwanted 3mux processes running")
			fmt.Printf("2. Run `rm -rf %s`\n\n", sessionInfo.path)
			fmt.Printf("Please also report this to %s\n", BUG_REPORT_URL)
			os.Exit(1)
		}
		fmt.Println("Session sucessfully killed.")
	case "ls", "ps":
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
	case "detach":
		if parentSessionID == "" {
			fmt.Println("Must be within session to detach")
			os.Exit(1)
		}
		detach(parentSessionID)
	default:
		fmt.Println(helpText)
		os.Exit(1)
	}
}

type SessionInfo struct {
	name           string
	uuid           string
	path           string
	fdPath         string
	killClientPath string
	killServerPath string
	detachPath     string
	resizePath     string
	logsPath       string
}

func detach(parentSessionID string) {
	net.Dial("unix", elaborateSessionInfo("", parentSessionID).detachPath)
}

func elaborateSessionInfo(name string, uuid string) *SessionInfo {
	dirPath := path.Join(threemuxDir, uuid)
	return &SessionInfo{
		name:           name,
		uuid:           uuid,
		path:           dirPath,
		fdPath:         path.Join(dirPath, "fd.sock"),
		killClientPath: path.Join(dirPath, "kill-client.sock"),
		killServerPath: path.Join(dirPath, "kill-server.sock"),
		detachPath:     path.Join(dirPath, "detach-server.sock"),
		resizePath:     path.Join(dirPath, "resize.sock"),
		logsPath:       path.Join(dirPath, "logs-server.txt"),
	}
}

func initializeSession(sessionName string) *SessionInfo {
	sessionUUID, err := randomIdentifier()
	if err != nil {
		fmt.Println("Session creation failed. Error while generating session UUID:", err)
		fmt.Printf("Please report this to %s\n", BUG_REPORT_URL)
		os.Exit(1)
	}

	sessionPath := path.Join(threemuxDir, sessionUUID)

	err = os.Mkdir(sessionPath, 0700)
	if err != nil {
		fmt.Println("Session creation failed. Error while creating session data directory:", err)
		fmt.Printf("Please report this to %s\n", BUG_REPORT_URL)
		os.Exit(1)
	}

	nameFilePath := path.Join(sessionPath, "name")
	err = ioutil.WriteFile(nameFilePath, []byte(sessionName), 0600)
	if err != nil {
		fmt.Println("Session creation failed. Error while recording session metadata:", err)
		err = os.Remove(sessionPath)
		if err != nil {
			fmt.Printf(
				"Session clean-up failed. Error while removing directory `%s`: %s",
				sessionPath, err,
			)
		}
		fmt.Printf("Please report this to %s\n", BUG_REPORT_URL)
		os.Exit(1)
	}

	return elaborateSessionInfo(sessionName, sessionUUID)
}

func findSession(sessionName string) (sessionInfo *SessionInfo, found bool, err error) {
	children, err := ioutil.ReadDir(threemuxDir)
	if err != nil {
		return nil, false, err
	}

	for _, child := range children {
		uuid := child.Name()
		dirPath := path.Join(threemuxDir, uuid)
		namePath := path.Join(dirPath, "name")
		nameRaw, err := ioutil.ReadFile(namePath)
		if err != nil {
			return nil, false, err
		}
		nameCleaned := strings.TrimSpace(string(nameRaw))

		if nameCleaned == sessionName {
			return elaborateSessionInfo(sessionName, uuid), true, nil
		}
	}

	return nil, false, nil
}

func refuseNesting() {
	fmt.Println("Refusing to run 3mux inside itself.")
	fmt.Println("If you want to do so anyway, `unset THREEMUX`.")
	os.Exit(1)
}

// returns empty name upon Ctrl-C
func defaultPrompt() (sessionName string, isNew bool) {
	os.MkdirAll(threemuxDir, 0755)
	children, err := ioutil.ReadDir(threemuxDir)
	if err != nil {
		panic(err)
	}

	optionIdx := 0
	options := []string{}
	idxToDir := map[int]string{}

	for idx, child := range children {
		dir := path.Join(threemuxDir, child.Name())
		thisName, _ := ioutil.ReadFile(path.Join(dir, "name"))

		options = append(options, strings.TrimSpace(string(thisName)))

		idxToDir[idx] = dir
	}

	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}
	defer terminal.Restore(0, oldState)
	defer fmt.Print("\x1b[m")

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

// log2(36^16) = 82 bits of entropy should be enough
const randIdentChars = "abcdefghijklmnopqrstuvwxyz0123456789"
const randIdentLen = 16

func randomIdentifier() (string, error) {
	out := make([]byte, randIdentLen)
	for i := 0; i < randIdentLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(randIdentChars))))
		if err != nil {
			return "", fmt.Errorf("cannot get random number: %w", err)
		}
		out[i] = randIdentChars[int(n.Int64())]
	}
	return string(out), nil
}
