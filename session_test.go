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
	print("test")

	bc.Write(`1 + 1`)
	bc.Write(`1 + 1`)
	time.Sleep(time.Second)
	bc.Write(`quit`)
	if o := <-bc.Output; o != "2" {
		t.Fail()
	}
	time.Sleep(time.Second * 3)

}

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
	// Output:
	// 2
}

// ExampleSessionWithOutput shows how to start a command that only runs
// for one second before being killed.
func ExampleNewSessionWithTimeout() {

	bc, err := interactive.NewSessionWithTimeout("bc", []string{}, time.Duration(time.Second))
	if err != nil {
		log.Fatal(err)
	}

	go outputPrinter(bc.Output)
	time.Sleep(time.Second)
	bc.Write(`1 + 1`) // this will never happen and there will not be any output
	// Output:
}

func outputPrinter(c chan string) {
	for s := range c {
		fmt.Println(s)
	}
}
