package main

import (
	"encoding/json"
	"io"
)

// jsonReport es el schema documentado del output --format=json.
// Estable para integraciones de terceros (dashboards, herramientas externas).
type jsonReport struct {
	ScannedFiles int             `json:"scannedFiles"`
	MatchCount   int             `json:"matchCount"`
	Matches      []jsonMatch     `json:"matches"`
	Summary      jsonReportStats `json:"summary"`
}

type jsonMatch struct {
	FileA      string `json:"fileA"`
	StartLineA int    `json:"startLineA"`
	EndLineA   int    `json:"endLineA"`
	FileB      string `json:"fileB"`
	StartLineB int    `json:"startLineB"`
	EndLineB   int    `json:"endLineB"`
	Tokens     int    `json:"tokens"`
}

type jsonReportStats struct {
	TopDuplicatedFiles []FileMatchCount `json:"topDuplicatedFiles"`
}

// reportJSON serializa el reporte a JSON. matches=nil produce
// arrays/objetos vacíos pero JSON válido. Encoding indented (2 espacios)
// para legibilidad humana.
func reportJSON(matches []Match, opts ReportOpts, w io.Writer) int {
	jms := make([]jsonMatch, 0, len(matches))
	for _, m := range matches {
		jms = append(jms, jsonMatch{
			FileA: m.FileA, StartLineA: m.StartLineA, EndLineA: m.EndLineA,
			FileB: m.FileB, StartLineB: m.StartLineB, EndLineB: m.EndLineB,
			Tokens: m.Tokens,
		})
	}
	rep := jsonReport{
		ScannedFiles: opts.ScannedCount,
		MatchCount:   len(matches),
		Matches:      jms,
		Summary:      jsonReportStats{TopDuplicatedFiles: topDuplicatedFiles(matches, 5)},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		return 0
	}
	return len(matches)
}
