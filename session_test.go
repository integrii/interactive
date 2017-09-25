package interactive_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/integrii/interactive"
)

func TestInteractiveCommandBC(t *testing.T) {
	// Run BC and hand it 1+1 to see what it does
	bc, err := interactive.NewInteractiveSession("bc", []string{})
	if err != nil {
		t.Fatal(err)
	}

	go outputPrinter(bc.Output)
	time.Sleep(time.Second)
	bc.Write(`1 + 1`)
	time.Sleep(time.Second)
}

func outputPrinter(c chan string) {
	for s := range c {
		fmt.Println(s)
	}
}

func ExampleSession() {
	// Start the command "bc" (a CLI calculator)
	bc, err := interactive.NewInteractiveSession("bc", []string{})
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
