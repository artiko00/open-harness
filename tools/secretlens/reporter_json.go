package main

import (
	"encoding/json"
	"io"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// jsonReport es el schema del output --format=json. Estable para
// integraciones externas (dashboards, CI). El sub-objeto "skipped"
// replica el contrato de dupelens: [{path, reason}].
type jsonReport struct {
	ScannedFiles int           `json:"scannedFiles"`
	FindingCount int           `json:"findingCount"`
	Findings     []jsonFinding `json:"findings"`
	Skipped      []jsonSkip    `json:"skipped"`
}

type jsonFinding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
}

// jsonSkip es un archivo no analizado, con el motivo de la omisión.
type jsonSkip struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// reportJSON serializa el reporte a JSON (indentado, sin ANSI) sobre w.
// findings/skips=nil producen arrays vacíos pero JSON válido.
func reportJSON(findings []Finding, skips []pathmatch.Skip, scanned int, w io.Writer) int {
	sortFindings(findings)
	jf := make([]jsonFinding, 0, len(findings))
	for _, f := range findings {
		jf = append(jf, jsonFinding{Path: f.RelPath, Line: f.Line, Rule: f.RuleName, Severity: f.Severity})
	}
	js := make([]jsonSkip, 0, len(skips))
	for _, s := range skips {
		js = append(js, jsonSkip{Path: s.Path, Reason: s.Reason})
	}
	rep := jsonReport{ScannedFiles: scanned, FindingCount: len(findings), Findings: jf, Skipped: js}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		return 0
	}
	return len(findings)
}
