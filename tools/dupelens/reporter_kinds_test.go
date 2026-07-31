package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func kindSample() []Match {
	return []Match{
		{FileA: "a.go", StartLineA: 1, EndLineA: 9, FileB: "b.go", StartLineB: 1, EndLineB: 9, Tokens: 60, Kind: "exact"},
		{FileA: "c.ts", StartLineA: 3, EndLineA: 8, FileB: "d.ts", StartLineB: 3, EndLineB: 8, Tokens: 55, Kind: "renamed"},
		{FileA: "e.ts", StartLineA: 3, EndLineA: 8, FileB: "f.ts", StartLineB: 3, EndLineB: 8, Tokens: 55, Kind: "renamed"},
		{FileA: "g.ts", StartLineA: 3, EndLineA: 8, FileB: "h.ts", StartLineB: 3, EndLineB: 8, Tokens: 55, Kind: "renamed"},
	}
}

func TestCountByKind(t *testing.T) {
	exact, renamed := countByKind(kindSample())
	if exact != 1 || renamed != 3 {
		t.Errorf("countByKind = (%d, %d); want (1, 3)", exact, renamed)
	}
}

func TestReportConsole_headerShowsKindBreakdown(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(kindSample(), ReportOpts{NoColor: true, ScannedCount: 8}, &buf)
	out := buf.String()
	header := strings.SplitN(out, "\n", 2)[0]
	if !strings.Contains(header, "1 exact") || !strings.Contains(header, "3 renamed") {
		t.Errorf("el encabezado debe desglosar los tipos, got %q", header)
	}
}

func TestReportConsole_summaryShowsKindBreakdown(t *testing.T) {
	var buf bytes.Buffer
	reportConsole(kindSample(), ReportOpts{NoColor: true, ScannedCount: 8}, &buf)
	last := lastLine(buf.String())
	if !strings.Contains(last, "1 exact") || !strings.Contains(last, "3 renamed") {
		t.Errorf("el SUMMARY debe desglosar los tipos, got %q", last)
	}
}

func TestReportJSON_includesKindCounts(t *testing.T) {
	var buf bytes.Buffer
	reportJSON(kindSample(), ReportOpts{ScannedCount: 8}, &buf)
	var rep struct {
		ExactCount   int `json:"exactCount"`
		RenamedCount int `json:"renamedCount"`
	}
	if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if rep.ExactCount != 1 || rep.RenamedCount != 3 {
		t.Errorf("exactCount=%d renamedCount=%d; want 1 y 3", rep.ExactCount, rep.RenamedCount)
	}
}
