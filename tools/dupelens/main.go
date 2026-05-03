package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

var osExit = os.Exit

func main() {
	osExit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 1
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:])
	case "init":
		return runInit(args[1:])
	case "version", "--version", "-v":
		fmt.Printf("dupelens %s\n", version)
		return 0
	case "help", "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 1
	}
}
