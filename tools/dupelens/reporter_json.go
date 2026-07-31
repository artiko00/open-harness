package main

import (
	"encoding/json"
	"io"
)

// jsonReport es el schema documentado del output --format=json.
// Estable para integraciones de terceros (dashboards, herramientas externas).
type jsonReport struct {
	ScannedFiles int `json:"scannedFiles"`
	MatchCount   int `json:"matchCount"`
	// ExactCount y RenamedCount desglosan MatchCount: --fail solo mira los
	// exact por defecto, así que el total por sí solo no dice si el gate rompe.
	ExactCount   int             `json:"exactCount"`
	RenamedCount int             `json:"renamedCount"`
	Matches      []jsonMatch     `json:"matches"`
	Skipped      []jsonSkip      `json:"skipped"`
	Summary      jsonReportStats `json:"summary"`
}

// jsonSkip es un archivo no analizado, con el motivo de la omisión.
type jsonSkip struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type jsonMatch struct {
	FileA      string `json:"fileA"`
	StartLineA int    `json:"startLineA"`
	EndLineA   int    `json:"endLineA"`
	FileB      string `json:"fileB"`
	StartLineB int    `json:"startLineB"`
	EndLineB   int    `json:"endLineB"`
	Tokens     int    `json:"tokens"`
	Kind       string `json:"kind"`
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
			Tokens: m.Tokens, Kind: m.Kind,
		})
	}
	jss := make([]jsonSkip, 0, len(opts.Skips))
	for _, s := range opts.Skips {
		jss = append(jss, jsonSkip{Path: s.Path, Reason: s.Reason})
	}
	exact, renamed := countByKind(matches)
	rep := jsonReport{
		ScannedFiles: opts.ScannedCount,
		MatchCount:   len(matches),
		ExactCount:   exact,
		RenamedCount: renamed,
		Matches:      jms,
		Skipped:      jss,
		Summary:      jsonReportStats{TopDuplicatedFiles: topDuplicatedFiles(matches, 5)},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		return 0
	}
	return len(matches)
}
