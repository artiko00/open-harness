package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	language string
	root     string
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	language := fs.String("lang", "auto", "language: go, typescript, python, ruby, rust, java, kotlin, csharp, auto")
	root := fs.String("dir", ".", "directory to scan")
	fail := fs.Bool("fail", false, "exit 1 if untested files found")
	fs.Parse(args)

	cfg := config{language: *language, root: *root}
	
	violations, err := checkCoverage(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if violations > 0 {
		fmt.Printf("\n%d file(s) without tests\n", violations)
		if *fail {
			os.Exit(1)
		}
	} else {
		fmt.Println("All source files have tests")
	}
}