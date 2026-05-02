package main

import (
	"fmt"
	"sort"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func report(results []FileResult, noColor bool) (violations int) {
	var v []FileResult
	for _, r := range results {
		if r.IsViolation() {
			v = append(v, r)
		}
	}

	sort.Slice(v, func(i, j int) bool {
		return v[i].Lines > v[j].Lines
	})

	if len(v) == 0 {
		if noColor {
			fmt.Printf("OK: all %d files within limits\n", len(results))
		} else {
			fmt.Printf("%s%sOK%s: all %d files within limits\n",
				colorBold, colorGreen, colorReset, len(results))
		}
		return 0
	}

	if noColor {
		fmt.Printf("VIOLATIONS (%d files exceed limits):\n\n", len(v))
	} else {
		fmt.Printf("%s%sVIOLATIONS%s (%d files exceed limits):\n\n",
			colorBold, colorRed, colorReset, len(v))
	}

	maxPathLen := 0
	for _, r := range v {
		if len(r.RelPath) > maxPathLen {
			maxPathLen = len(r.RelPath)
		}
	}
	if maxPathLen > 60 {
		maxPathLen = 60
	}

	for _, r := range v {
		path := r.RelPath
		if len(path) > 60 {
			path = "..." + path[len(path)-57:]
		}

		excess := r.Lines - r.MaxLines
		indicator := "▲"

		if noColor {
			fmt.Printf("  %-*s  %4d lines  (max: %d, +%d)\n",
				maxPathLen, path, r.Lines, r.MaxLines, excess)
		} else {
			lineColor := colorYellow
			if excess > r.MaxLines/2 {
				lineColor = colorRed
			}
			fmt.Printf("  %s%-*s%s  %s%s%4d lines%s  %s(max: %d, %s%d)%s\n",
				colorBold, maxPathLen, path, colorReset,
				colorBold, lineColor, r.Lines, colorReset,
				colorGray, r.MaxLines, indicator, excess, colorReset)
		}
	}

	fmt.Printf("\nSUMMARY: %d violation(s) in %d file(s) scanned\n", len(v), len(results))
	return len(v)
}
