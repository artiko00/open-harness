package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.3.1"

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
		fmt.Printf("secretlens %s\n", version)
		return 0
	case "help", "--help", "-h":
		printUsage()
		return 0
	case "--tutorial", "tutorial":
		printTutorial(hasNoColor(args[1:]))
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 1
	}
}

// hasNoColor devuelve true si "--no-color" aparece en args.
func hasNoColor(args []string) bool {
	for _, a := range args {
		if a == "--no-color" {
			return true
		}
	}
	return false
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "secretlens.json", "output file name")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists, not overwriting\n", *output)
		return 1
	}

	if err := os.WriteFile(*output, []byte(defaultConfigJSON()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing config: %v\n", err)
		return 1
	}

	fmt.Printf("created %s\n", *output)
	return 0
}
