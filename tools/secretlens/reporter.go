package main

import (
	"fmt"
	"sort"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

var severityOrder = map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}

func report(findings []Finding, skips []pathmatch.Skip, noColor bool) int {
	if len(findings) > 0 {
		sortFindings(findings)
		printFindings(findings, noColor)
	}
	printSkipped(skips, noColor)

	if len(findings) == 0 && !pathmatch.AnyFailsGate(skips) {
		printOK(len(skips), noColor)
		return 0
	}

	fmt.Printf("SUMMARY: %d potential secret(s) found — review before pushing%s\n",
		len(findings), sufijoResumen(len(skips)))
	return len(findings)
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		si := severityOrder[findings[i].Severity]
		sj := severityOrder[findings[j].Severity]
		if si != sj {
			return si < sj
		}
		if findings[i].RelPath != findings[j].RelPath {
			return findings[i].RelPath < findings[j].RelPath
		}
		return findings[i].Line < findings[j].Line
	})
}

func printFindings(findings []Finding, noColor bool) {
	if noColor {
		fmt.Printf("SECRETS FOUND (%d finding(s)):\n\n", len(findings))
	} else {
		fmt.Printf("%s%sSECRETS FOUND%s (%d finding(s)):\n\n",
			colorBold, colorRed, colorReset, len(findings))
	}

	for _, f := range findings {
		if noColor {
			fmt.Printf("  [%s] %s:%d\n    Rule: %s\n    Line: %s\n\n",
				f.Severity, f.RelPath, f.Line, f.RuleName, f.Content)
			continue
		}
		sColor := severityColor(f.Severity)
		fmt.Printf("  %s[%s]%s %s%s%s:%s%d%s\n    %sRule:%s %s\n    %sLine:%s %s\n\n",
			sColor, f.Severity, colorReset,
			colorBold, f.RelPath, colorReset,
			colorCyan, f.Line, colorReset,
			colorGray, colorReset, f.RuleName,
			colorGray, colorReset, f.Content)
	}
}

func severityColor(s string) string {
	switch s {
	case "critical":
		return colorRed + colorBold
	case "high":
		return colorRed
	case "medium":
		return colorYellow
	default:
		return colorGray
	}
}
