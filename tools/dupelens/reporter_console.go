package main

import (
	"fmt"
	"io"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// reportConsole imprime matches con formato legible en consola, seguido de la
// sección SKIPPED y de la línea final (OK o SUMMARY). Si opts.Snippet está
// provisto, agrega 3 líneas de contexto por archivo.
func reportConsole(matches []Match, opts ReportOpts, w io.Writer) int {
	if len(matches) > 0 {
		writeMatches(matches, opts, w)
		fmt.Fprintln(w)
	}
	writeSkipped(w, opts.Skips)

	if len(matches) == 0 && !pathmatch.AnyFailsGate(opts.Skips) {
		fmt.Fprintf(w, "%s%sOK%s: no duplicate code blocks found  (%d files scanned)%s\n",
			boldOpt(opts.NoColor), greenOpt(opts.NoColor), resetOpt(opts.NoColor),
			opts.ScannedCount, skippedSuffix(opts.Skips, " (%d skipped)"))
		return 0
	}

	if top := topDuplicatedFiles(matches, 5); len(top) > 0 {
		fmt.Fprintln(w, "Top duplicated files:")
		for _, t := range top {
			fmt.Fprintf(w, "  - %s  (%d match(es))\n", t.File, t.Count)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "SUMMARY: %d match(es)%s across %d files%s\n",
		len(matches), kindBreakdown(matches), opts.ScannedCount,
		skippedSuffix(opts.Skips, ", %d skipped"))
	return len(matches)
}

// kindBreakdown formatea el desglose exact/renamed entre paréntesis, o cadena
// vacía si no hay matches.
func kindBreakdown(matches []Match) string {
	if len(matches) == 0 {
		return ""
	}
	exact, renamed := countByKind(matches)
	return fmt.Sprintf(" (%d exact · %d renamed)", exact, renamed)
}

// writeMatches imprime el encabezado DUPLICATES y el detalle de cada match.
func writeMatches(matches []Match, opts ReportOpts, w io.Writer) {
	fmt.Fprintf(w, "%s%sDUPLICATES%s (%d match(es)%s found in %d files):\n\n",
		boldOpt(opts.NoColor), redOpt(opts.NoColor), resetOpt(opts.NoColor),
		len(matches), kindBreakdown(matches), opts.ScannedCount)

	for _, m := range matches {
		fmt.Fprintf(w, "  %s%s:%d-%d%s  ↔  %s%s:%d-%d%s  %s(%d tokens, %s)%s\n",
			boldOpt(opts.NoColor), m.FileA, m.StartLineA, m.EndLineA, resetOpt(opts.NoColor),
			boldOpt(opts.NoColor), m.FileB, m.StartLineB, m.EndLineB, resetOpt(opts.NoColor),
			grayOpt(opts.NoColor), m.Tokens, m.Kind, resetOpt(opts.NoColor))
		writeSnippetPair(w, opts, m)
	}
}

// writeSnippetPair imprime cada fragmento del match bajo su propia ruta (no
// concatenados), para que se vea a qué archivo pertenece cada snippet.
func writeSnippetPair(w io.Writer, opts ReportOpts, m Match) {
	if opts.Snippet == nil {
		return
	}
	fmt.Fprintf(w, "  %s%s:%d%s\n", grayOpt(opts.NoColor), m.FileA, m.StartLineA, resetOpt(opts.NoColor))
	fmt.Fprint(w, opts.Snippet(m.FileA, m.StartLineA, 3))
	fmt.Fprintf(w, "  %s%s:%d%s\n", grayOpt(opts.NoColor), m.FileB, m.StartLineB, resetOpt(opts.NoColor))
	fmt.Fprint(w, opts.Snippet(m.FileB, m.StartLineB, 3))
}
