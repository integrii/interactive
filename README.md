# interactive 🐚
Interactive is a package for easily executing and interacting with CLI commands using channels.  Uses a PTY and simple channels for inputting lines of strings and reading lines of strings.  You can always go direct to the running process and write bytes, though, too.  A cool example of this in use is my [headlessChrome](https://github.com/integrii/headlessChrome) package which starts headless chrome with `--repl` (a command line javascript console interface) for very, very realistic website automation.

Automate any command line task with a go program!  Special thanks to [github.com/kr/pty](https://github.com/kr/pty)

## Get It

`go get -u github.com/integrii/interactive`

## Read the Godoc

[https://godoc.org/github.com/integrii/interactive](https://godoc.org/github.com/integrii/interactive)


## Example

```go

func main() {
  // Start the command "bc" (a CLI calculator)
  bc, err := interactive.NewSession("bc", []string{})
  if err != nil {
    panic(err)
  }

  // start a concurrent output reader from the output channel of our command
  go outputPrinter(bc.Output)

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

You can run and control nearly anything you could from your console this way.
