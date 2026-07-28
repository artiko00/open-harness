package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// reportConsole imprime el reporte legible: encabezado y lista de archivos sin
// test (si los hay), la sección SKIPPED y la línea final (OK/SUMMARY).
func reportConsole(cov coverage, noColor bool) {
	if len(cov.untested) > 0 {
		fmt.Println(untestedHeader(len(cov.untested), noColor))
		suffix := "- no test found"
		if cov.pkg {
			suffix = "- no tests found"
		}
		for _, p := range cov.untested {
			fmt.Println(violationLine(p, suffix, noColor))
		}
	}
	printSkipped(cov.skips)
	fmt.Println(summaryLine(len(cov.untested), cov.skips, noColor))
}

// jsonSkip es un archivo no analizado, con el motivo de la omisión.
type jsonSkip struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// jsonReport es el schema del output --format=json de testlens.
type jsonReport struct {
	SourceFiles   int        `json:"sourceFiles"`
	UntestedCount int        `json:"untestedCount"`
	Untested      []string   `json:"untested"`
	Skipped       []jsonSkip `json:"skipped"`
}

// reportJSON serializa el reporte a JSON válido (arrays vacíos, nunca null).
func reportJSON(cov coverage, w io.Writer) {
	skips := make([]jsonSkip, 0, len(cov.skips))
	for _, s := range cov.skips {
		skips = append(skips, jsonSkip{Path: s.Path, Reason: s.Reason})
	}
	untested := cov.untested
	if untested == nil {
		untested = []string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(jsonReport{
		SourceFiles:   cov.scanned,
		UntestedCount: len(cov.untested),
		Untested:      untested,
		Skipped:       skips,
	})
}
