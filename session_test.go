package interactive_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/integrii/interactive"
)

func TestInteractiveCommandBC(t *testing.T) {

	// interactive.Debug = true

	// Run BC and hand it 1+1 to see what it does
	bc, err := interactive.NewSession("bc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	bc.Write(`1 + 1`)
	bc.Write(`quit`)
	time.Sleep(time.Second)
	outputPrinter(bc.Output)

}

// ExampleNewSession shows how to make a new interactive session
// with a command.  Output comes from the string channel bc.Output
// while iput is passed in with the bc.Write func.  Note that we
// named the interactive session "bc" here because we're running bc.
func ExampleNewSession() {
	// Start the command "bc" (a CLI calculator)
	bc, err := interactive.NewSession("bc", []string{})
	if err != nil {
		panic(err)
	}

	// start a concurrent output reader from the output channel of our command
	go func(outChan chan string) {
		for s := range outChan {
			fmt.Println(s)
		}
	}(bc.Output)

	// wait a second for the process to init
	time.Sleep(time.Second)

	// write 1 + 1 to the bc prompt
	bc.Write(`1 + 1`)

	// wait one second for the output to come and be displayed
	time.Sleep(time.Second)
}

// ExampleSessionWithOutput shows how to start a command that only runs
// for one second before being killed.  The 1 + 1 operation never happens
// becauase the command is killed prior to its running
func ExampleNewSessionWithTimeout() {

	bc, err := interactive.NewSessionWithTimeout("bc", []string{}, time.Duration(time.Second))
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 2)
	bc.Write(`1 + 1`) // this will never happen and there will not be any output
}

func outputPrinter(c chan string) {
	for s := range c {
		fmt.Println(s)
	}
}
