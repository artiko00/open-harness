package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "linelens.json", "path to config file")
	maxLines := fs.Int("max", 0, "override max lines for all files")
	failOnViolation := fs.Bool("fail", false, "exit with code 1 if violations found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	format := fs.String("format", "console", "output format: console | json")
	root := fs.String("dir", ".", "directory to scan")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *format != "console" && *format != "json" {
		fmt.Fprintf(os.Stderr, "invalid format %q (valid: console, json)\n", *format)
		return 1
	}

	maxSet := false
	configSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "max":
			maxSet = true
		case "config":
			configSet = true
		}
	})
	if maxSet && *maxLines <= 0 {
		fmt.Fprintf(os.Stderr, "max must be greater than 0\n")
		return 1
	}

	if configSet {
		if _, err := os.Stat(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "config file %q not found\n", *configPath)
			return 1
		}
	}

	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "directory %q not accessible: %v\n", *root, err)
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	results, skips, err := scan(*root, cfg, *maxLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		return 1
	}

	violations := report(results, skips, *format, *noColor, os.Stdout)
	if *failOnViolation && (violations > 0 || pathmatch.AnyFailsGate(skips)) {
		return 1
	}
	return 0
}
