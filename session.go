package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/kr/pty"
)

// Debug enables debug output for this package to console
var Debug bool

// Session is an interactive console session for the specified
// command and arguments.
type Session struct {
	StdIn   io.Writer   // input to be written to the console
	StdOut  io.Reader   // output coming from the console
	StdErr  io.Reader   // error output from the shell
	Input   chan string // incoming lines of input
	Output  chan string // outgoing lines of input
	Cmd     *exec.Cmd   // cmd that holds this cmd instance
	outDone bool
	timeout time.Duration
	pty     *os.File
}

// startOutputReader reads pty output and puts it into
// the output channel one line at a time
func (i *Session) startOutputReader() {
	reader := bufio.NewScanner(i.pty)
	debug("Output reader looking for output")
	for reader.Scan() {
		text := reader.Text()
		debug("Output reader got text:", text)
		i.Output <- text
		debug("Reader passed text to outut channel:", text)
	}
	debug("stdout done")
	i.outDone = true
}

// startInputForwarder starts a forwarder of input channel
// to running session
func (i *Session) startInputForwarder() {
	debug("stdin forwarder running")
	for l := range i.Input {
		debug("input channel request to write string:", l)
		i.writeString(l)
	}
	debug("stdin done")
}

// Write writes an output line into the session
func (i *Session) Write(s string) {
	debug("Writing", s, "to input channel")
	i.Input <- s
}

// WriteString writes a string to the console as if you wrote
// it and pressed enter.
func (i *Session) writeString(s string) error {
	debug("Writing string:", s)
	s = s + "\n"
	_, err := i.pty.Write([]byte(s))
	return err
}

// Exit exits the running command and closes the input channel
func (i *Session) Exit() {
	i.Cmd.Process.Signal(os.Interrupt)
}

// Init runs things required to initalize a session.
// No need to call outside of NewInteractiveSession (which does
// it for you)
func (i *Session) Init() error {

	// kick off the command and ensure it closes when done
	var err error

	i.pty, err = pty.Start(i.Cmd)
	if err != nil {
		return err
	}

	debug("Spawned command as PID", i.Cmd.Process.Pid)

	go i.startOutputReader()
	go i.startInputForwarder()
	go i.cleanupWhenDone()

	return nil

}

// cleanupWhenDone cleans up channels when done or kills
// the process if it runs too long
func (i *Session) cleanupWhenDone() {
	debug("Waiting for session to complete.")

	// start sesson and send error to a done channel
	// so we can select timeout or channel return later
	done := make(chan error)
	go func() { done <- i.Cmd.Wait() }()

	// if the timeout is greater than 0, start a timer for it
	if i.timeout > 0 {
		// kill the cmd if it goes too long
		select {
		case err := <-done:
			debug("Session command has ended.", err)
		case <-time.After(i.timeout):
			// timed out. force close.
			debug("Session execution timed out.  Closing forcefully.")
			i.ForceClose()
		}
	}

	// wait for all readers to complete reads before closing channels
	debug("Waiting for output scanner to exit.")
	for !i.outDone {
		time.Sleep(time.Millisecond) // helps reduce cpu use
	}

	// indicate the session is closed and close our channels
	debug("Command exited. Closing input and output channels.")

	// close our output channel to cause upstream channel readers
	// to complete work
	close(i.Output)

}

// NewSessionWithTimeout starts a new session but kills it
// if it runs longer than the specified timeout.  Pass  0
// for no timeout or use NewSession()
func NewSessionWithTimeout(command string, args []string, timeout time.Duration) (*Session, error) {

	var session Session
	var err error

	// assign the timeout to the struct
	session.timeout = timeout

	// make channels for input and outut communication to the process
	session.Input = make(chan string, 1)
	session.Output = make(chan string, 5000)

	// setup the command and input/output pipes
	session.Cmd = exec.Command(command, args...)

	session.StdIn, err = session.Cmd.StdinPipe()
	if err != nil {
		return &session, err
	}

	debug("Starting command:", session.Cmd.Args)

	// start channeling output and other requirements
	err = session.Init()

	// command is online and healthy, return to the user
	return &session, err
}

// NewSession starts a new interactive command session
func NewSession(command string, args []string) (*Session, error) {
	return NewSessionWithTimeout(command, args, 0)
}

// ForceClose issues a force kill to the command (SIGKILL)
func (i *Session) ForceClose() {
	i.Cmd.Process.Kill()
}

func debug(s ...interface{}) {
	if Debug {
		fmt.Println(s...)
	}
}
