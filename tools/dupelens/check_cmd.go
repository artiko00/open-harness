package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "dupelens.json", "path to config file")
	minTokens := fs.Int("min-tokens", 0, "override min duplicate token threshold")
	failOnViolation := fs.Bool("fail", false, "exit with code 1 if duplicates found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	format := fs.String("format", "console", "output format: console | json")
	root := fs.String("dir", ".", "directory to scan")
	verbose := fs.Bool("verbose", false, "print timing and progress to stderr")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "directory %q not accessible: %v\n", *root, err)
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	t0 := time.Now()
	matches, scanned, err := scan(*root, cfg, *minTokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		return 1
	}
	tScan := time.Since(t0)

	opts := ReportOpts{Format: *format, NoColor: *noColor, ScannedCount: scanned}
	if *format != "json" {
		opts.Snippet = makeSnippetFunc(*root)
	}
	violations := report(matches, opts, os.Stdout)
	if *verbose {
		fmt.Fprintf(os.Stderr, "[verbose] scan=%s (%d files, %d matches), total=%s\n",
			tScan, scanned, len(matches), time.Since(t0))
	}
	if *failOnViolation && violations > 0 {
		return 1
	}
	return 0
}
