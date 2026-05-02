package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		runCheck(os.Args[2:])
	case "init":
		runInit(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("testlens %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "testlens.json", "output file name")
	fs.Parse(args)

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists, not overwriting\n", *output)
		os.Exit(1)
	}

	config := `{
  "language": "auto",
  "skip": ["node_modules", ".git", "vendor", "dist", "build"],
  "languages": {
    "go": { "extensions": [".go"], "testSuffixes": ["_test"] },
    "typescript": { "extensions": [".ts", ".tsx"], "testSuffixes": [".test", ".spec"] }
  }
}`
	if err := os.WriteFile(*output, []byte(config), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", *output)
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