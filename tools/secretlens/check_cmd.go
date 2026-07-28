package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "secretlens.json", "path to config file")
	failOnFinding := fs.Bool("fail", false, "exit with code 1 if secrets found")
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

	if _, err := os.Stat(*root); err != nil {
		fmt.Fprintf(os.Stderr, "directory %q not accessible: %v\n", *root, err)
		return 1
	}

	explicitConfig := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			explicitConfig = true
		}
	})
	if explicitConfig {
		if _, err := os.Stat(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "config file %q not found\n", *configPath)
			return 1
		}
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	findings, skips, scanned, err := scan(*root, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		return 1
	}

	var count int
	if *format == "json" {
		count = reportJSON(findings, skips, scanned, os.Stdout)
	} else {
		count = report(findings, skips, *noColor)
	}
	if *failOnFinding && (count > 0 || pathmatch.AnyFailsGate(skips)) {
		return 1
	}
	return 0
}
