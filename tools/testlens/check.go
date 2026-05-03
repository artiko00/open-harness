package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	language string
	root     string
	fail     bool
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	language := fs.String("lang", "auto", "language: go, typescript, python, ruby, rust, java, kotlin, csharp, auto")
	root := fs.String("dir", ".", "directory to scan")
	fail := fs.Bool("fail", false, "exit 1 if untested files found")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		return 1
	}

	cfg := config{language: *language, root: *root, fail: *fail}
	
	violations, err := checkCoverage(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if violations > 0 {
		fmt.Printf("\n%d file(s) without tests\n", violations)
		if *fail {
			return 1
		}
	} else {
		fmt.Println("All source files have tests")
	}
	return 0
}