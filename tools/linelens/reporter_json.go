package main

import (
	"encoding/json"
	"io"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// jsonReport es el schema del output --format=json de linelens.
type jsonReport struct {
	ScannedFiles   int             `json:"scannedFiles"`
	ViolationCount int             `json:"violationCount"`
	Violations     []jsonViolation `json:"violations"`
	Skipped        []jsonSkip      `json:"skipped"`
}

// jsonViolation es un archivo que excede su límite de líneas o de anidamiento.
// `lines` son líneas de CÓDIGO; `physical` es el total físico del archivo.
type jsonViolation struct {
	Path       string `json:"path"`
	Lines      int    `json:"lines"`
	Physical   int    `json:"physical"`
	MaxLines   int    `json:"maxLines"`
	Nesting    int    `json:"nesting"`
	MaxNesting int    `json:"maxNesting"`
}

// jsonSkip es un archivo no analizado, con el motivo de la omisión.
type jsonSkip struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// reportJSON serializa el reporte a JSON indentado (2 espacios) en w.
// results vacío produce arrays vacíos pero JSON válido.
func reportJSON(results []FileResult, skips []pathmatch.Skip, w io.Writer) int {
	jvs := make([]jsonViolation, 0)
	for _, r := range results {
		if r.IsViolation() {
			jvs = append(jvs, jsonViolation{
				Path:       r.RelPath,
				Lines:      r.Lines,
				Physical:   r.Physical,
				MaxLines:   r.MaxLines,
				Nesting:    r.Nesting,
				MaxNesting: r.MaxNesting,
			})
		}
	}
	jss := make([]jsonSkip, 0, len(skips))
	for _, s := range skips {
		jss = append(jss, jsonSkip{Path: s.Path, Reason: s.Reason})
	}
	rep := jsonReport{
		ScannedFiles:   len(results),
		ViolationCount: len(jvs),
		Violations:     jvs,
		Skipped:        jss,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		return 0
	}
	return len(jvs)
}
