package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0-scaffold"

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
		fmt.Printf("dupelens %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	configPath := fs.String("config", "dupelens.json", "path to config file")
	minTokens := fs.Int("min-tokens", 0, "override min duplicate token threshold")
	failOnViolation := fs.Bool("fail", false, "exit with code 1 if duplicates found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	format := fs.String("format", "console", "output format: console | json")
	root := fs.String("dir", ".", "directory to scan")
	fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	matches, scanned, err := scan(*root, cfg, *minTokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	opts := ReportOpts{
		Format:       *format,
		NoColor:      *noColor,
		ScannedCount: scanned,
	}
	if *format != "json" {
		opts.Snippet = makeSnippetFunc(*root)
	}
	violations := report(matches, opts, os.Stdout)
	if *failOnViolation && violations > 0 {
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "dupelens.json", "output file name")
	fs.Parse(args)

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists, not overwriting\n", *output)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, []byte(defaultConfigJSON()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", *output)
}
