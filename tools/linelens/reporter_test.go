package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func TestReport_Empty(t *testing.T) {
	if count := reportConsole([]FileResult{}, nil, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoViolationsNoColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "a.go", Lines: 50, MaxLines: 100},
		{RelPath: "b.go", Lines: 80, MaxLines: 100},
	}
	if count := reportConsole(results, nil, true); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_NoViolationsColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "a.go", Lines: 50, MaxLines: 100},
	}
	if count := reportConsole(results, nil, false); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestReport_ViolationsNoColor(t *testing.T) {
	results := []FileResult{
		{RelPath: "big.go", Lines: 150, MaxLines: 100},
		{RelPath: "bigger.go", Lines: 200, MaxLines: 100},
	}
	if count := reportConsole(results, nil, true); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestReport_ViolationsColorYellow(t *testing.T) {
	// excess=30, maxLines/2=50, 30 <= 50 → yellow
	results := []FileResult{{RelPath: "big.go", Lines: 130, MaxLines: 100}}
	if count := reportConsole(results, nil, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_ViolationsColorRed(t *testing.T) {
	// excess=60, maxLines/2=50, 60 > 50 → red
	results := []FileResult{{RelPath: "big.go", Lines: 160, MaxLines: 100}}
	if count := reportConsole(results, nil, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_LongPathNoColor(t *testing.T) {
	longPath := "src/very/deeply/nested/path/to/some/file/that/exceeds/sixty/chars.go"
	results := []FileResult{{RelPath: longPath, Lines: 150, MaxLines: 100}}
	if count := reportConsole(results, nil, true); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReport_LongPathColor(t *testing.T) {
	longPath := "src/very/deeply/nested/path/to/some/file/that/exceeds/sixty/chars.go"
	results := []FileResult{{RelPath: longPath, Lines: 160, MaxLines: 100}}
	if count := reportConsole(results, nil, false); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestReportJSON_EmptyProducesValidSchema(t *testing.T) {
	var buf bytes.Buffer
	exit := reportJSON([]FileResult{{RelPath: "a.go", Lines: 10, MaxLines: 100}}, nil, &buf)
	if exit != 0 {
		t.Errorf("sin violations debe retornar 0, got %d", exit)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v\nbuf: %s", err, buf.String())
	}
	if parsed["scannedFiles"].(float64) != 1 {
		t.Errorf("scannedFiles incorrecto: %v", parsed["scannedFiles"])
	}
	if parsed["violationCount"].(float64) != 0 {
		t.Errorf("violationCount debe ser 0, got %v", parsed["violationCount"])
	}
	if _, ok := parsed["skipped"]; !ok {
		t.Error("debe incluir la clave skipped")
	}
}

func TestReportJSON_WithViolationsAndSkips(t *testing.T) {
	var buf bytes.Buffer
	results := []FileResult{
		{RelPath: "big.go", Lines: 150, MaxLines: 100},
		{RelPath: "ok.go", Lines: 10, MaxLines: 100},
	}
	skips := []pathmatch.Skip{{Path: "bin.dat", Reason: pathmatch.ReasonBinary}}
	exit := reportJSON(results, skips, &buf)
	if exit != 1 {
		t.Errorf("con 1 violation debe retornar 1, got %d", exit)
	}
	var parsed struct {
		ScannedFiles   int `json:"scannedFiles"`
		ViolationCount int `json:"violationCount"`
		Violations     []struct {
			Path     string `json:"path"`
			Lines    int    `json:"lines"`
			MaxLines int    `json:"maxLines"`
		} `json:"violations"`
		Skipped []struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		} `json:"skipped"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if parsed.ViolationCount != 1 || len(parsed.Violations) != 1 {
		t.Fatalf("violations incorrecto: %+v", parsed)
	}
	if parsed.Violations[0].Path != "big.go" || parsed.Violations[0].Lines != 150 {
		t.Errorf("violation incorrecta: %+v", parsed.Violations[0])
	}
	if len(parsed.Skipped) != 1 || parsed.Skipped[0].Path != "bin.dat" {
		t.Errorf("skipped incorrecto: %+v", parsed.Skipped)
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

func TestReportJSON_WriteError(t *testing.T) {
	if exit := reportJSON(nil, nil, errorWriter{}); exit != 0 {
		t.Errorf("reportJSON con write error = %d, want 0", exit)
	}
}

func TestReport_DispatchByFormat(t *testing.T) {
	var buf bytes.Buffer
	results := []FileResult{{RelPath: "big.go", Lines: 150, MaxLines: 100}}
	if n := report(results, nil, "json", true, &buf); n != 1 {
		t.Errorf("dispatch json = %d, want 1", n)
	}
	if err := json.Unmarshal(buf.Bytes(), &map[string]any{}); err != nil {
		t.Errorf("dispatch json no produjo JSON válido: %v", err)
	}
	// console no escribe en w; retorna el conteo de violations.
	var buf2 bytes.Buffer
	if n := report(results, nil, "console", true, &buf2); n != 1 {
		t.Errorf("dispatch console = %d, want 1", n)
	}
}
