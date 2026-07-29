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

	if r.MaxLines > 0 {
		fmt.Fprintf(w, "\nSUMMARY: %d files, %d lines counted, %d excluded; limit %d files / %d lines (%s)\n",
			r.Countable, r.Lines, len(r.Excluded), r.Max, r.MaxLines, r.Mode)
		return
	}
	fmt.Fprintf(w, "\nSUMMARY: %d counted, %d excluded, limit %d\n",
		r.Countable, len(r.Excluded), r.Max)
}

// metrics arma la parte "N files (max M)[, L lines (max K)]" del estado.
func metrics(r report) string {
	s := fmt.Sprintf("%d files (max %d)", r.Countable, r.Max)
	if r.MaxLines > 0 {
		s += fmt.Sprintf(", %d lines (max %d)", r.Lines, r.MaxLines)
	}
	return s
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
		fmt.Fprintf(w, "  %s: %s\n", label, metrics(r))
		return
	}
	fmt.Fprintf(w, "  %s%s%s%s: %s\n", colorBold, color, label, colorReset, metrics(r))
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
