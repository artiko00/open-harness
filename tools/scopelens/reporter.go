package main

import (
	"fmt"
	"io"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorBold  = "\033[1m"
)

// printReport emite el encabezado, el estado, el desglose por categoría y la
// línea SUMMARY. Con noColor no escribe ninguna secuencia ANSI.
func printReport(r report, noColor bool, w io.Writer) {
	fmt.Fprintf(w, "scopelens %s — %s\n\n", version, header(r))
	printStatus(r, noColor, w)

	printFiles(w, catSource, r.Source)
	printFiles(w, catTest, r.Test)
	printExcluded(w, r.Excluded)

	fmt.Fprintf(w, "\nSUMMARY: %d counted, %d excluded, limit %d\n",
		r.Countable, len(r.Excluded), r.Max)
}

func header(r report) string {
	if r.MergeBase == "" {
		return fmt.Sprintf("%s (sólo staged)", r.Branch)
	}
	return fmt.Sprintf("%s vs %s (merge-base %s)", r.Branch, r.Base, r.MergeBase)
}

func printStatus(r report, noColor bool, w io.Writer) {
	label, color := "OK", colorGreen
	if r.exceeded() {
		label, color = "FAIL", colorRed
	}
	if noColor {
		fmt.Fprintf(w, "  %s: %d files (max %d)\n", label, r.Countable, r.Max)
		return
	}
	fmt.Fprintf(w, "  %s%s%s%s: %d files (max %d)\n",
		colorBold, color, label, colorReset, r.Countable, r.Max)
}

func printFiles(w io.Writer, cat string, files []string) {
	if len(files) == 0 {
		return
	}
	fmt.Fprintf(w, "\n  %s (%d)\n", cat, len(files))
	for _, f := range files {
		fmt.Fprintf(w, "    %s\n", f)
	}
}

func printExcluded(w io.Writer, files []excludedFile) {
	if len(files) == 0 {
		return
	}
	ancho := 0
	for _, f := range files {
		if len(f.Path) > ancho {
			ancho = len(f.Path)
		}
	}
	fmt.Fprintf(w, "\n  %s (%d)\n", catExcluded, len(files))
	for _, f := range files {
		fmt.Fprintf(w, "    %-*s    %s\n", ancho, f.Path, f.Reason)
	}
}
