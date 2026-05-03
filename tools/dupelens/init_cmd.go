package main

import (
	"flag"
	"fmt"
	"os"
)

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "dupelens.json", "output file name")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if _, err := os.Stat(*output); err == nil {
		fmt.Fprintf(os.Stderr, "%s already exists, not overwriting (delete it first or use --output)\n", *output)
		return 1
	}

	if err := os.WriteFile(*output, []byte(defaultConfigJSON()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing config to %q: %v\n", *output, err)
		return 1
	}

	fmt.Printf("created %s — edit thresholds and re-run 'dupelens check'\n", *output)
	return 0
}
