package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ViolationResult represents the outcome of checking a single file
type ViolationResult struct {
	Path    string
	HasTest bool
	TestFor string
}

// printViolations prints the violations in a human-readable format
func printViolations(results []ViolationResult, root string) int {
	violations := 0
	for _, r := range results {
		if !r.HasTest {
			relPath, _ := filepath.Rel(root, r.Path)
			fmt.Printf("  %s - no test found\n", relPath)
			violations++
		}
	}
	return violations
}

// printSummary prints the final summary message
func printSummary(violations int) {
	if violations > 0 {
		fmt.Printf("\n%d file(s) without tests\n", violations)
	} else {
		fmt.Println("All source files have tests")
	}
}

// exitWithCode exits with appropriate exit code based on fail flag and violations
func exitWithCode(fail bool, violations int) {
	if fail && violations > 0 {
		os.Exit(1)
	}
}