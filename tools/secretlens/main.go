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
		fmt.Printf("secretlens %s\n", version)
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
	configPath := fs.String("config", "secretlens.json", "path to config file")
	failOnFinding := fs.Bool("fail", false, "exit with code 1 if secrets found")
	noColor := fs.Bool("no-color", false, "disable colored output")
	root := fs.String("dir", ".", "directory to scan")
	fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	findings, err := scan(*root, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	count := report(findings, *noColor)
	if *failOnFinding && count > 0 {
		os.Exit(1)
	}
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "secretlens.json", "output file name")
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
