package main

import "fmt"

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
	colorBold  = "\033[1m"
)

// violationLine formats a single "path - reason" line, colored unless noColor.
func violationLine(relPath, suffix string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("  %s %s", relPath, suffix)
	}
	return fmt.Sprintf("  %s%s%s %s%s%s", colorBold, relPath, colorReset, colorGray, suffix, colorReset)
}

// summaryLine formats the final summary, colored unless noColor.
func summaryLine(violations int, noColor bool) string {
	if violations > 0 {
		if noColor {
			return fmt.Sprintf("\n%d file(s) without tests", violations)
		}
		return fmt.Sprintf("\n%s%s%d file(s) without tests%s", colorBold, colorRed, violations, colorReset)
	}
	if noColor {
		return "All source files have tests"
	}
	return fmt.Sprintf("%s%sAll source files have tests%s", colorBold, colorGreen, colorReset)
}
