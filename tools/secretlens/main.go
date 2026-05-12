package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.2"

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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 1
	}
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "secretlens.json", "path to config file")
	failOnFinding := fs.Bool("fail", false, "exit with code 1 if secrets found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	root := fs.String("dir", ".", "directory to scan")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	findings, err := scan(*root, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		return 1
	}

	count := report(findings, *noColor)
	if *failOnFinding && count > 0 {
		return 1
	}
	return 0
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
