package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.2.5"

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
		fmt.Printf("testlens %s\n", version)
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

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "testlens.json", "output file name")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists, not overwriting\n", *output)
		return 1
	}

	config := `{
  "language": "auto",
  "exclude": ["node_modules", ".git", "vendor", "dist", "build", "testdata"]
}`
	if err := os.WriteFile(*output, []byte(config), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing config: %v\n", err)
		return 1
	}

	fmt.Printf("created %s\n", *output)
	return 0
}

func printUsage() {
	fmt.Print(`testlens — find source files missing tests

USAGE:
  testlens <command> [options]

COMMANDS:
  check     scan directories and report files without tests
  init      create a default testlens.json config
  version   show version

CHECK OPTIONS:
  --lang      language or 'auto' for detection (default: auto)
  --dir       directory to scan                   (default: .)
  --fail      exit 1 if untested files found     (for CI)
  --no-color  disable colored output

LANGUAGES:
  go, typescript, javascript, python, ruby, rust, java, kotlin, csharp

EXAMPLES:
  testlens check
  testlens check --fail
  testlens check --lang typescript --dir ./src
  testlens check --lang go

EXIT CODES:
  0 - all files have tests (or --fail not used and violations found)
  1 - violations found when --fail is set
`)
}