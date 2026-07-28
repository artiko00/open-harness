package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// escribeCfg crea un linelens.json con maxLines=100 en dir y devuelve su ruta.
func escribeCfg(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "linelens.json")
	os.WriteFile(p, []byte(`{"default":{"maxLines":100}}`), 0644)
	return p
}

// capturaStdout ejecuta fn redirigiendo os.Stdout y devuelve lo escrito.
func capturaStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRunCheck_FormatJSON_ValidStdout(t *testing.T) {
	dir := t.TempDir()
	cfg := escribeCfg(t, dir)
	os.WriteFile(filepath.Join(dir, "big.go"), []byte(strings.Repeat("x\n", 200)), 0644)
	os.WriteFile(filepath.Join(dir, "ok.go"), []byte("package main\n"), 0644)

	var code int
	out := capturaStdout(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", cfg, "--format", "json"})
	})
	if code != 0 {
		t.Fatalf("esperado 0, got %d", code)
	}
	var parsed struct {
		ScannedFiles int `json:"scannedFiles"`
		Violations   []struct {
			Path string `json:"path"`
		} `json:"violations"`
		Skipped []struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		} `json:"skipped"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("stdout no es JSON válido: %v\nout: %s", err, out)
	}
	if parsed.ScannedFiles < 1 {
		t.Errorf("scannedFiles debe ser >=1, got %d", parsed.ScannedFiles)
	}
	if len(parsed.Violations) < 1 {
		t.Errorf("debe haber al menos 1 violation, got %+v", parsed.Violations)
	}
}

func TestRunCheck_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	var code int
	out := capturaStdout(t, func() {
		code = runCheck([]string{"--dir", dir, "--format", "xml"})
	})
	if code != 1 {
		t.Errorf("formato inválido debe dar exit 1, got %d", code)
	}
	if strings.Contains(out, "OK:") || strings.Contains(out, "SUMMARY:") {
		t.Errorf("stdout no debe llevar reporte con formato inválido: %s", out)
	}
}

func TestRunCheck_DirNotAccessible(t *testing.T) {
	if code := runCheck([]string{"--dir", "/no/existe/xyz123"}); code != 1 {
		t.Errorf("dir inexistente debe dar exit 1, got %d", code)
	}
}

func TestRunCheck_ScanErrorAfterStat(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no confiable como root")
	}
	dir := t.TempDir()
	cfg := escribeCfg(t, t.TempDir()) // config válido en un dir legible aparte
	// El directorio a escanear existe (pasa Stat) pero es ilegible (el walk falla).
	os.Chmod(dir, 0000)
	defer os.Chmod(dir, 0755)
	if code := runCheck([]string{"--dir", dir, "--config", cfg}); code != 1 {
		t.Errorf("scan error debe dar exit 1, got %d", code)
	}
}

func TestRunCheck_MaxNonPositive(t *testing.T) {
	dir := t.TempDir()
	if code := runCheck([]string{"--dir", dir, "--max", "-5"}); code != 1 {
		t.Errorf("--max -5 debe dar exit 1, got %d", code)
	}
	if code := runCheck([]string{"--dir", dir, "--max", "0"}); code != 1 {
		t.Errorf("--max 0 explícito debe dar exit 1, got %d", code)
	}
}

func TestRunCheck_MaxPositiveApplies(t *testing.T) {
	dir := t.TempDir()
	cfg := escribeCfg(t, dir)
	os.WriteFile(filepath.Join(dir, "f.go"), []byte(strings.Repeat("x\n", 20)), 0644)
	if code := runCheck([]string{"--dir", dir, "--config", cfg, "--max", "10", "--fail"}); code != 1 {
		t.Errorf("--max 10 con archivo de 20 líneas debe violar, got %d", code)
	}
}

func TestRunCheck_LastLineOK(t *testing.T) {
	dir := t.TempDir()
	cfg := escribeCfg(t, dir)
	os.WriteFile(filepath.Join(dir, "ok.go"), []byte("package main\n"), 0644)
	out := capturaStdout(t, func() {
		runCheck([]string{"--dir", dir, "--config", cfg, "--no-color"})
	})
	last := ultimaLinea(out)
	if !strings.HasPrefix(last, "OK:") {
		t.Errorf("última línea debe empezar con 'OK:', got %q", last)
	}
}

func TestRunCheck_LastLineSummary(t *testing.T) {
	dir := t.TempDir()
	cfg := escribeCfg(t, dir)
	os.WriteFile(filepath.Join(dir, "big.go"), []byte(strings.Repeat("x\n", 200)), 0644)
	out := capturaStdout(t, func() {
		runCheck([]string{"--dir", dir, "--config", cfg, "--no-color"})
	})
	if !strings.Contains(out, "VIOLATIONS (") {
		t.Errorf("debe haber encabezado VIOLATIONS, got: %s", out)
	}
	last := ultimaLinea(out)
	if !strings.HasPrefix(last, "SUMMARY:") {
		t.Errorf("última línea debe empezar con 'SUMMARY:', got %q", last)
	}
}

// ultimaLinea devuelve la última línea no vacía de s.
func ultimaLinea(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}
