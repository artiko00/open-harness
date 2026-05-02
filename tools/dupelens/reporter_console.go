package main

import (
	"fmt"
	"io"
	"sort"
)

// reportConsole imprime matches con formato legible en consola.
// Si opts.Snippet está provisto, agrega 3 líneas de contexto por archivo.
func reportConsole(matches []Match, opts ReportOpts, w io.Writer) int {
	if len(matches) == 0 {
		fmt.Fprintf(w, "%s%sOK%s: no duplicate code blocks found  (%d files scanned)\n",
			boldOpt(opts.NoColor), greenOpt(opts.NoColor), resetOpt(opts.NoColor), opts.ScannedCount)
		return 0
	}

	fmt.Fprintf(w, "%s%sDUPLICATES%s (%d match(es) found in %d files):\n\n",
		boldOpt(opts.NoColor), redOpt(opts.NoColor), resetOpt(opts.NoColor),
		len(matches), opts.ScannedCount)

	for _, m := range matches {
		fmt.Fprintf(w, "  %s%s:%d-%d%s  ↔  %s%s:%d-%d%s  %s(%d tokens)%s\n",
			boldOpt(opts.NoColor), m.FileA, m.StartLineA, m.EndLineA, resetOpt(opts.NoColor),
			boldOpt(opts.NoColor), m.FileB, m.StartLineB, m.EndLineB, resetOpt(opts.NoColor),
			grayOpt(opts.NoColor), m.Tokens, resetOpt(opts.NoColor))
		if opts.Snippet != nil {
			fmt.Fprint(w, opts.Snippet(m.FileA, m.StartLineA, 3))
			if m.FileB != m.FileA {
				fmt.Fprint(w, opts.Snippet(m.FileB, m.StartLineB, 3))
			}
		}
	}

	top := topDuplicatedFiles(matches, 5)
	fmt.Fprintf(w, "\nSUMMARY: %d match(es) across %d files\n", len(matches), opts.ScannedCount)
	if len(top) > 0 {
		fmt.Fprintln(w, "Top duplicated files:")
		for _, t := range top {
			fmt.Fprintf(w, "  - %s  (%d match(es))\n", t.File, t.Count)
		}
	}
	return len(matches)
}

// FileMatchCount es par usado en summary "top duplicated files".
type FileMatchCount struct {
	File  string `json:"file"`
	Count int    `json:"count"`
}

// topDuplicatedFiles cuenta apariciones de cada archivo en el set de matches
// (sumando FileA y FileB) y retorna los N con más matches, ordenado descendente.
func topDuplicatedFiles(matches []Match, n int) []FileMatchCount {
	counts := make(map[string]int)
	for _, m := range matches {
		counts[m.FileA]++
		if m.FileB != m.FileA {
			counts[m.FileB]++
		}
	}
	out := make([]FileMatchCount, 0, len(counts))
	for f, c := range counts {
		out = append(out, FileMatchCount{File: f, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].File < out[j].File
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
