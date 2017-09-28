package interactive

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	StdIn      io.Writer   // input to be written to the console
	StdOut     io.Reader   // output coming from the console
	StdErr     io.Reader   // error output from the shell
	Input      chan string // incoming lines of input
	Output     chan string // outgoing lines of input
	Cmd        *exec.Cmd   // cmd that holds this cmd instance
	PTY        *os.File    // the tty for the session
	stdOutDone bool
	stdInDone  bool
	stdErrdone bool
}

// WriteString writes a string to the console as if you wrote
// it and pressed enter.
func (i *Session) writeString(s string) error {
	if Debug {
		fmt.Println("Writing string:", s)
	}
	_, err := i.PTY.WriteString(s + "\r")
	return err
}

// startErrorReader starts an error reader that outputs
// to the output channel
func (i *Session) startErrorReader() {
	reader := bufio.NewScanner(i.StdErr)
	if Debug {
		fmt.Println("Error reader looking for output")
	}
	for reader.Scan() {
		text := reader.Text()
		if Debug {
			fmt.Println("Error reader got text:", text)
		}
		i.Output <- text
		if Debug {
			fmt.Println("Error reader passed output to channel:", text)
		}
	}

	// safely flag this as done
	i.stdErrdone = true
}

// startOutputReader reads output and puts it into the output channel
func (i *Session) startOutputReader() {
	reader := bufio.NewScanner(i.PTY)
	if Debug {
		fmt.Println("Output reader looking for output")
	}
	for reader.Scan() {
		text := reader.Text()
		if Debug {
			fmt.Println("Output reader got text:", text)
		}
		i.Output <- text
		if Debug {
			fmt.Println("Reader passed text to outut channel:", text)
		}
	}
	i.stdOutDone = true
}

// startInputForwarder starts a forwarder of input channel
// to running session
func (i *Session) startInputForwarder() {
	for l := range i.Input {
		if Debug {
			fmt.Println("Got request to write string:", l)
		}
		i.writeString(l)
	}
	i.stdInDone = true
}

// Exit exits the running command and closes the input channel
func (i *Session) Exit() {
	i.Cmd.Process.Signal(os.Interrupt)
}

// Init runs things required to initalize a session.
// No need to call outside of NewInteractiveSession (which does
// it for you)
func (i *Session) Init(timeout time.Duration) error {
	// kick off the command and ensure it closes when done
	var err error
	i.PTY, err = pty.Start(i.Cmd)
	if err != nil {
		return err
	}

	if Debug {
		fmt.Println("Spawned command as PID", i.Cmd.Process.Pid)
	}

	go i.startOutputReader()
	go i.startErrorReader()
	go i.startInputForwarder()
	go i.closeWhenCompleted(timeout)

	return nil
}

// Write writes an output line into the session
func (i *Session) Write(s string) {

	// dont actually write if the command has completed
	i.Input <- s
}

// closeWhenCompleted closes ouput channels to cause readers to
// end gracefully when the command completes
func (i *Session) closeWhenCompleted(timeout time.Duration) {

	if Debug {
		fmt.Println("Waiting for session to complete.")
	}

	// if the timeout is greater than 0, start a timer for it
	if timeout > 0 {

		// kill the cmd if it goes too long
		done := make(chan error)
		go func() { done <- i.Cmd.Wait() }()
		select {
		case err := <-done:
			// exited on its own
			if err != nil {
				log.Println("Interactive session exited with error:", err)
			}
		case <-time.After(timeout):
			// timed out. force close.
			i.ForceClose()
		}
	} else {
		// if no timeout is specified, just start the command
		err := i.Cmd.Wait()
		if err != nil {
			log.Println("Interactive session exited with error:", err)
		}
	}

	// wait for all readers to complete reads before closing channels
	for !i.stdErrdone || !i.stdInDone || !i.stdOutDone {
		time.Sleep(time.Millisecond)
	}

	// indicate the session is closed and close our channels
	if Debug {
		fmt.Println("Command exited. Closing all channels.")
	}

	// close our channels to cause channel readers to complete work
	close(i.Input)
	close(i.Output)

}

// NewSessionWithTimeout starts a new session but kills it
// if it runs longer than the specified timeout.  Pass  0
// for no timeout or use NewSession()
func NewSessionWithTimeout(command string, args []string, timeout time.Duration) (*Session, error) {

	var session Session
	var err error

	// setup the command and input/output pipes
	session.Cmd = exec.Command(command, args...)
	errPipe, err := session.Cmd.StderrPipe()
	if err != nil {
		return &session, err
	}
	inPipe, err := session.Cmd.StdinPipe()
	if err != nil {
		return &session, err
	}
	outPipe, err := session.Cmd.StdoutPipe()
	if err != nil {
		return &session, err
	}

	// bind sessions to struct
	session.StdOut = outPipe
	session.StdIn = inPipe
	session.StdErr = errPipe

	// make channels for input and outut communication to the process
	session.Input = make(chan string, 1)
	session.Output = make(chan string, 5000)

	if Debug {
		fmt.Println("Starting command:", session.Cmd.Args)
	}

	// start channeling output and other requirements
	err = session.Init(timeout)

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
