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
