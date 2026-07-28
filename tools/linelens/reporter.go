package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// report despacha al backend según format: json va al io.Writer w (sin ANSI,
// testeable sin capturar el fd global); console imprime a stdout.
func report(results []FileResult, skips []pathmatch.Skip, format string, noColor bool, w io.Writer) int {
	if format == "json" {
		return reportJSON(results, skips, w)
	}
	return reportConsole(results, skips, noColor)
}

func reportConsole(results []FileResult, skips []pathmatch.Skip, noColor bool) (violations int) {
	var v []FileResult
	for _, r := range results {
		if r.IsViolation() {
			v = append(v, r)
		}
	}

	sort.Slice(v, func(i, j int) bool {
		return v[i].Lines > v[j].Lines
	})

	if len(v) > 0 {
		printViolations(v, noColor)
	}
	printSkips(skips, noColor)

	if len(v) == 0 && !pathmatch.AnyFailsGate(skips) {
		printOK(len(results), len(skips), noColor)
		return 0
	}

	fmt.Printf("\nSUMMARY: %d violation(s) in %d file(s) scanned%s\n",
		len(v), len(results), sufijoSkipped(len(skips)))
	return len(v)
}

