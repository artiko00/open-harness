package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func sampleMatches() []Match {
	return []Match{
		{FileA: "a.go", StartLineA: 1, EndLineA: 5, FileB: "b.go", StartLineB: 10, EndLineB: 14, Tokens: 5},
		{FileA: "a.go", StartLineA: 20, EndLineA: 25, FileB: "c.go", StartLineB: 1, EndLineB: 6, Tokens: 6},
	}
}

func TestReportJSON_emptyMatchesProducesValidSchema(t *testing.T) {
	var buf bytes.Buffer
	exit := reportJSON(nil, ReportOpts{ScannedCount: 42}, &buf)
	if exit != 0 {
		t.Errorf("empty matches must return 0, got %d", exit)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v\nbuf: %s", err, buf.String())
	}
	if parsed["scannedFiles"].(float64) != 42 {
		t.Errorf("scannedFiles incorrecto: %v", parsed["scannedFiles"])
	}
	if parsed["matchCount"].(float64) != 0 {
		t.Errorf("matchCount debe ser 0, got %v", parsed["matchCount"])
	}
}

func TestReportJSON_withMatches_includesAll(t *testing.T) {
	var buf bytes.Buffer
	matches := sampleMatches()
	reportJSON(matches, ReportOpts{ScannedCount: 10}, &buf)

	var parsed struct {
		ScannedFiles int `json:"scannedFiles"`
		MatchCount   int `json:"matchCount"`
		Matches      []struct {
			FileA      string `json:"fileA"`
			StartLineA int    `json:"startLineA"`
			Tokens     int    `json:"tokens"`
		} `json:"matches"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if parsed.MatchCount != 2 {
		t.Errorf("matchCount = %d; want 2", parsed.MatchCount)
	}
	if len(parsed.Matches) != 2 {
		t.Fatalf("matches len = %d; want 2", len(parsed.Matches))
	}
	if parsed.Matches[0].FileA != "a.go" || parsed.Matches[0].Tokens != 5 {
		t.Errorf("primer match incorrecto: %+v", parsed.Matches[0])
	}
}

func TestReportConsole_emptyMatches(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(nil, ReportOpts{ScannedCount: 50, NoColor: true}, &buf)
	out := buf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("output debe contener 'OK', got: %s", out)
	}
	if !strings.Contains(out, "50") {
		t.Errorf("output debe mencionar scannedFiles=50, got: %s", out)
	}
}

func TestReportConsole_withMatches_showsFilesAndRanges(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(sampleMatches(), ReportOpts{ScannedCount: 10, NoColor: true}, &buf)
	out := buf.String()
	for _, want := range []string{"a.go", "b.go", "c.go", "1-5", "10-14", "20-25"} {
		if !strings.Contains(out, want) {
			t.Errorf("output debe contener %q, got: %s", want, out)
		}
	}
}

func TestReportConsole_summaryShowsStats(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(sampleMatches(), ReportOpts{ScannedCount: 10, NoColor: true}, &buf)
	out := buf.String()
	if !strings.Contains(out, "2 ") && !strings.Contains(out, "2\n") {
		t.Errorf("summary debe mostrar conteo de matches, got: %s", out)
	}
}

func TestReportConsole_withSnippetFunc_showsContext(t *testing.T) {
	var buf bytes.Buffer
	snippet := func(file string, startLine, count int) string {
		return "  | snippet_for_" + file + "_at_line_" + itoa(startLine) + "\n"
	}
	opts := ReportOpts{ScannedCount: 1, NoColor: true, Snippet: snippet}
	reportConsole(sampleMatches(), opts, &buf)
	out := buf.String()
	if !strings.Contains(out, "snippet_for_a.go_at_line_1") {
		t.Errorf("snippet de a.go debe aparecer, got: %s", out)
	}
	if !strings.Contains(out, "snippet_for_b.go_at_line_10") {
		t.Errorf("snippet de b.go debe aparecer, got: %s", out)
	}
}

func TestReportConsole_withoutSnippetFunc_noSnippets(t *testing.T) {
	var buf bytes.Buffer
	opts := ReportOpts{ScannedCount: 1, NoColor: true}
	reportConsole(sampleMatches(), opts, &buf)
	out := buf.String()
	if strings.Contains(out, "snippet_") {
		t.Errorf("sin SnippetFunc no deben aparecer snippets, got: %s", out)
	}
}

func TestReport_dispatchSelectsByFormat(t *testing.T) {
	var bufJSON, bufConsole bytes.Buffer
	report(sampleMatches(), ReportOpts{Format: "json", ScannedCount: 5}, &bufJSON)
	report(sampleMatches(), ReportOpts{Format: "console", ScannedCount: 5, NoColor: true}, &bufConsole)

	// JSON debe parsear; console no.
	if err := json.Unmarshal(bufJSON.Bytes(), &map[string]any{}); err != nil {
		t.Errorf("dispatch JSON no produjo JSON válido: %v", err)
	}
	if !strings.Contains(bufConsole.String(), "a.go") {
		t.Errorf("dispatch console no produjo formato esperado")
	}
}

// itoa local para evitar dep de strconv en tests donde queremos simpleza.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
