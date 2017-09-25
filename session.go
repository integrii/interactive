package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/kr/pty"
)

// Debug enables debug output for this package to console
var Debug bool

// Session is an interactive console session for the specified
// command and arguments.
type Session struct {
	stdIn   io.Writer   // input to be written to the console
	stdOut  io.Reader   // output coming from the console
	stdErr  io.Reader   // error output from the shell
	Input   chan string // incoming lines of input
	Output  chan string // outgoing lines of input
	cmd     *exec.Cmd   // cmd that holds this chrome instance
	pty     *os.File    // the tty for the session
	command string      // command to run
	args    []string    // arguments to pass to running command
}

// WriteString writes a string to the console as if you wrote
// it and pressed enter.
func (i *Session) writeString(s string) error {
	if Debug {
		fmt.Println("Writing string:", s)
	}
	_, err := i.pty.WriteString(s + "\r")
	return err
}

// startErrorReader starts an error reader that outputs
// to the output channel
func (i *Session) startErrorReader() {
	reader := bufio.NewScanner(i.stdErr)
	if Debug {
		fmt.Println("Error reader looking for output")
	}
	for reader.Scan() {
		if Debug {
			fmt.Println("Error reader got text:", reader.Text())
		}
		i.Output <- reader.Text()
		if Debug {
			fmt.Println("Error reader passed output to channel:", reader.Text())
		}
	}
}

// startOutputReader reads output and puts it into the output channel
func (i *Session) startOutputReader() {
	reader := bufio.NewScanner(i.pty)
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
}

// Exit exits the running command and closes the input channel
func (i *Session) Exit() {

	i.cmd.Process.Signal(os.Interrupt)

	// close will cause the io workers to stop gracefully
	close(i.Input)
}

// Init runs things required to initalize a chrome session.
// No need to call outside of NewInteractiveSession (which does
// it for you)
func (i *Session) Init() {
	go i.startOutputReader()
	go i.startErrorReader()
	go i.startInputForwarder()
	go i.closeWhenCompleted()
}

// Write writes an output line into the session
func (i *Session) Write(s string) {
	i.Input <- s
}

// closeWhenCompleted closes ouput channels to cause readers to
// end gracefully when the command completes
func (i *Session) closeWhenCompleted() {

	if Debug {
		fmt.Println("Spawned chrome as PID", i.cmd.Process.Pid)
	}

	i.cmd.Wait()
	if Debug {
		fmt.Println("Command exited. Closing channels.")
	}
	close(i.Output)

	i.forceClose() // when complete, make sure the PID dies (chrome never does on its own as of writing)
}

// NewInteractiveSession starts a new chrome headless session.
func NewInteractiveSession(command string, args []string) (*Session, error) {
	var session Session
	var err error

	// setup the command and input/output pipes
	session.cmd = exec.Command(command, args...)
	errPipe, err := session.cmd.StderrPipe()
	if err != nil {
		return &session, err
	}
	inPipe, err := session.cmd.StdinPipe()
	if err != nil {
		return &session, err
	}
	outPipe, err := session.cmd.StdoutPipe()
	if err != nil {
		return &session, err
	}

	// bind sessions to struct
	session.stdOut = outPipe
	session.stdIn = inPipe
	session.stdErr = errPipe

	// make channels for input and outut communication to the process
	session.Input = make(chan string, 1)
	session.Output = make(chan string, 5000)

	if Debug {
		fmt.Println("Starting command:", session.cmd.Args)
	}

	// kick off the command and ensure it closes when done
	session.pty, _ = pty.Start(session.cmd)
	if err != nil {
		return &session, err
	}

	// start channeling output and other requirements
	session.Init()

	// command is online and healthy, return to the user
	return &session, err

}

// forceClose issues a force kill to the command
func (i *Session) forceClose() {
	i.cmd.Process.Kill()
}
