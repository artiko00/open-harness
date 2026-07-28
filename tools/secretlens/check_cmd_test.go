package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// capturarStdoutStderr ejecuta fn redirigiendo stdout y stderr por separado.
func capturarStdoutStderr(t *testing.T, fn func()) (string, string) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	ro, wo, _ := os.Pipe()
	re, we, _ := os.Pipe()
	os.Stdout, os.Stderr = wo, we
	fn()
	wo.Close()
	we.Close()
	os.Stdout, os.Stderr = origOut, origErr

	var so, se strings.Builder
	copyPipe(&so, ro)
	copyPipe(&se, re)
	return so.String(), se.String()
}

func copyPipe(sb *strings.Builder, r *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	r.Close()
}

func TestRunCheck_FormatJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.go"), []byte("const k = \"AKIAQNB7Q7AIIVSMBGPF\"\n"), 0644)

	var code int
	out, _ := capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--format", "json"})
	})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(out, "\033[") {
		t.Errorf("json no debe llevar ANSI, got:\n%s", out)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("stdout no es JSON válido: %v\n%s", err, out)
	}
	for _, k := range []string{"scannedFiles", "findingCount", "findings", "skipped"} {
		if _, ok := m[k]; !ok {
			t.Errorf("falta la clave %q en el JSON", k)
		}
	}
	if m["scannedFiles"].(float64) != 1 {
		t.Errorf("scannedFiles = %v, want 1", m["scannedFiles"])
	}
	fs := m["findings"].([]any)
	if len(fs) != 1 {
		t.Fatalf("findings = %d, want 1", len(fs))
	}
	f := fs[0].(map[string]any)
	for _, k := range []string{"path", "line", "rule", "severity"} {
		if _, ok := f[k]; !ok {
			t.Errorf("falta la clave %q en el finding", k)
		}
	}
}

func TestRunCheck_FormatInvalido(t *testing.T) {
	var code int
	out, errOut := capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", ".", "--format", "xml"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "console, json") {
		t.Errorf("stderr debe mencionar los formatos válidos, got:\n%s", errOut)
	}
	if strings.Contains(out, "OK:") || strings.Contains(out, "SUMMARY:") {
		t.Errorf("stdout no debe llevar reporte, got:\n%s", out)
	}
}

func TestRunCheck_DirInexistente(t *testing.T) {
	var code int
	_, errOut := capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", "/no/existe/abc123", "--format", "console"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not accessible") {
		t.Errorf("stderr debe decir 'not accessible', got:\n%s", errOut)
	}
}

func TestRunCheck_ConsolaOK(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\n"), 0644)

	out, _ := capturarStdoutStderr(t, func() {
		runCheck([]string{"--dir", dir, "--no-color"})
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "OK:") {
		t.Errorf("última línea = %q, debe empezar con OK:", last)
	}
}

func TestRunCheck_ConsolaSummary(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.go"), []byte("const k = \"AKIAQNB7Q7AIIVSMBGPF\"\n"), 0644)

	out, _ := capturarStdoutStderr(t, func() {
		runCheck([]string{"--dir", dir, "--no-color"})
	})
	if !strings.Contains(out, "SECRETS FOUND") {
		t.Errorf("falta el encabezado de hallazgos, got:\n%s", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "SUMMARY:") {
		t.Errorf("última línea = %q, debe empezar con SUMMARY:", last)
	}
}

func TestRunCheck_ScanErrorConDirValido(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "secretlens.json")
	os.WriteFile(cfgPath, []byte(`{"patterns":[{"name":"bad","pattern":"[invalid","severity":"high"}]}`), 0644)

	var code int
	_, errOut := capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", cfgPath})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "scan error") {
		t.Errorf("stderr debe decir 'scan error', got:\n%s", errOut)
	}
}

// errWriter falla siempre para cubrir la rama de error del encoder JSON.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

func TestReportJSON_EncoderError(t *testing.T) {
	if n := reportJSON(nil, nil, 0, errWriter{}); n != 0 {
		t.Errorf("con error de encoder debe retornar 0, got %d", n)
	}
}

func TestReportJSON_Directo(t *testing.T) {
	findings := []Finding{{RelPath: "a.go", Line: 3, RuleName: "R", Severity: "high"}}
	skips := []pathmatch.Skip{{Path: "b.bin", Reason: pathmatch.ReasonBinary}}
	var buf bytes.Buffer
	n := reportJSON(findings, skips, 2, &buf)
	if n != 1 {
		t.Errorf("n = %d, want 1", n)
	}
	var rep jsonReport
	if err := json.Unmarshal(buf.Bytes(), &rep); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if rep.ScannedFiles != 2 || rep.FindingCount != 1 || len(rep.Skipped) != 1 {
		t.Errorf("reporte inesperado: %+v", rep)
	}
}
