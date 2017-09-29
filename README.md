# interactive 🐚
Interactive is a package for easily executing and interacting with running commands using channels.  Uses a PTY and simple channels for inputting lines of strings and reading lines of strings.  You can always go direct to the running process and write strings, though, too.

Automate nearly any command line execution in a go program!  Special thanks to [github.com/kr/pty](https://github.com/kr/pty)

## Get It

`go get -u github.com/integrii/interactive`

## Read the Godoc

[https://godoc.org/github.com/integrii/interactive](https://godoc.org/github.com/integrii/interactive)


## Example

```go

func main() {
  // Start the command "bc" (a CLI calculator)
  bc, err := interactive.NewInteractiveSession("bc", []string{})
  if err != nil {
    panic(err)
  }

  // start a concurrent output reader from the output channel of our command
  go outputPrinter(bc.Output)

  // wait a second for the process to init
  time.Sleep(time.Second)

  // write 1 + 1 to the bc prompt
  bc.Write(`1 + 1`)

  // wait one second for the output to come and be displayed
  time.Sleep(time.Second)

}

func outputPrinter(c chan string) {
  for s := range c {
    fmt.Println(s)
  }
}

```

This will print to the console the following:

```bash
bc 1.06
Copyright 1991-1994, 1997, 1998, 2000 Free Software Foundation, Inc.
This is free software with ABSOLUTELY NO WARRANTY.
For details type `warranty'.
1 + 1
2
```

You can run nearly anything you could from your console this way.
