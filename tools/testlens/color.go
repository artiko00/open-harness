package main

import (
	"fmt"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

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

// untestedHeader formats the header shown before the list of untested files.
func untestedHeader(n int, noColor bool) string {
	if noColor {
		return fmt.Sprintf("UNTESTED (%d file(s) without tests):", n)
	}
	return fmt.Sprintf("%s%sUNTESTED (%d file(s) without tests):%s", colorBold, colorRed, n, colorReset)
}

// summaryLine formats the final line del contrato: empieza con "OK:" cuando no
// hay violaciones ni omitidos que rompan el gate, o con "SUMMARY:" en otro caso.
// El conteo de archivos omitidos se informa siempre que haya alguno (tarea 2.9).
func summaryLine(untested int, skips []pathmatch.Skip, noColor bool) string {
	if untested == 0 && !pathmatch.AnyFailsGate(skips) {
		sufijo := ""
		if len(skips) > 0 {
			sufijo = fmt.Sprintf(" (%d skipped)", len(skips))
		}
		if noColor {
			return "OK: all source files have tests" + sufijo
		}
		return fmt.Sprintf("%s%sOK: all source files have tests%s%s", colorBold, colorGreen, sufijo, colorReset)
	}
	sufijo := ""
	if len(skips) > 0 {
		sufijo = fmt.Sprintf(", %d skipped", len(skips))
	}
	if noColor {
		return fmt.Sprintf("SUMMARY: %d file(s) without tests%s", untested, sufijo)
	}
	return fmt.Sprintf("%s%sSUMMARY: %d file(s) without tests%s%s", colorBold, colorRed, untested, sufijo, colorReset)
}
