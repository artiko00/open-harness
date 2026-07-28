package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// checkCoverageCounts adapta checkCoverage a la forma (violations, skips, err)
// que usan los tests preexistentes.
func checkCoverageCounts(cfg config) (int, []pathmatch.Skip, error) {
	cov, err := checkCoverage(cfg)
	return len(cov.untested), cov.skips, err
}

// capturaStderr ejecuta fn con stderr redirigido y devuelve lo impreso.
func capturaStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// creaConfigVacio deja un testlens.json valido y vacio en dir para que el
// contrato de --config explicito (debe existir) se cumpla en los tests.
func creaConfigVacio(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "testlens.json")
	os.WriteFile(path, []byte("{}"), 0644)
	return path
}

// ultimaLinea devuelve la última línea no vacía de s.
func ultimaLinea(s string) string {
	lineas := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lineas[len(lineas)-1]
}

func TestCheck_FormatJSONValido(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.ts"), []byte("export const x = 1"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	var code int
	out := capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "typescript", "--dir", dir, "--format", "json",
			"--config", cfgPath})
	})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.Contains(out, "\033") {
		t.Errorf("JSON no debe llevar códigos ANSI:\n%s", out)
	}

	var rep struct {
		SourceFiles   int      `json:"sourceFiles"`
		UntestedCount int      `json:"untestedCount"`
		Untested      []string `json:"untested"`
		Skipped       []struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		} `json:"skipped"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("JSON inválido: %v\n%s", err, out)
	}
	if rep.SourceFiles != 1 {
		t.Errorf("sourceFiles = %d, want 1", rep.SourceFiles)
	}
	if rep.UntestedCount != 1 || len(rep.Untested) != 1 || rep.Untested[0] != "foo.ts" {
		t.Errorf("untested = %v (count %d), want [foo.ts]", rep.Untested, rep.UntestedCount)
	}
	if rep.Skipped == nil {
		t.Error("skipped debe ser un array (aunque vacío), no null")
	}
}

func TestCheck_FormatInvalido(t *testing.T) {
	dir := t.TempDir()
	var code int
	out := capturaSalida(t, func() {
		errOut := capturaStderr(t, func() {
			code = runCheck([]string{"--dir", dir, "--format", "xml",
				"--config", filepath.Join(dir, "testlens.json")})
		})
		if !strings.Contains(errOut, "console, json") {
			t.Errorf("stderr debe mencionar los formatos válidos: %q", errOut)
		}
		if !strings.Contains(errOut, `invalid format "xml"`) {
			t.Errorf("stderr debe indicar el formato inválido: %q", errOut)
		}
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if strings.Contains(out, "OK:") || strings.Contains(out, "SUMMARY:") || strings.Contains(out, "UNTESTED") {
		t.Errorf("stdout no debe llevar reporte con formato inválido: %q", out)
	}
}

func TestCheck_DirInaccesible(t *testing.T) {
	var code int
	errOut := capturaStderr(t, func() {
		code = runCheck([]string{"--dir", "/no/existe/xyz123"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not accessible") {
		t.Errorf("stderr debe indicar directorio inaccesible: %q", errOut)
	}
}

func TestCheck_SinHallazgosOK(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "foo_test.py"), []byte("def test_foo(): pass"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	out := capturaSalida(t, func() {
		runCheck([]string{"--lang", "python", "--dir", dir, "--no-color",
			"--config", cfgPath})
	})
	if !strings.HasPrefix(ultimaLinea(out), "OK:") {
		t.Errorf("última línea debe empezar con OK:, got %q", ultimaLinea(out))
	}
}

func TestCheck_ConHallazgosSummary(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	out := capturaSalida(t, func() {
		runCheck([]string{"--lang", "python", "--dir", dir, "--no-color",
			"--config", cfgPath})
	})
	if !strings.Contains(out, "UNTESTED (1 file(s) without tests):") {
		t.Errorf("falta el encabezado UNTESTED en:\n%s", out)
	}
	if !strings.HasPrefix(ultimaLinea(out), "SUMMARY:") {
		t.Errorf("última línea debe empezar con SUMMARY:, got %q", ultimaLinea(out))
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("con hallazgos no debe aparecer OK: en:\n%s", out)
	}
}

func TestCheck_AvisoDeteccionVaAStderr(t *testing.T) {
	dir := t.TempDir()
	// Extensión desconocida: la detección falla y emite el aviso por stderr.
	os.WriteFile(filepath.Join(dir, "foo.xyz"), []byte("x = 1"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	var errOut string
	out := capturaSalida(t, func() {
		errOut = capturaStderr(t, func() {
			runCheck([]string{"--lang", "auto", "--dir", dir, "--no-color",
				"--config", cfgPath})
		})
	})
	if strings.Contains(out, "Could not detect language") {
		t.Errorf("el aviso de detección NO debe salir por stdout:\n%s", out)
	}
	if !strings.Contains(errOut, "Could not detect language") {
		t.Errorf("el aviso de detección debe ir a stderr: %q", errOut)
	}
}

func TestCheck_JSONSinViolaciones(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "foo_test.py"), []byte("def test_foo(): pass"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	out := capturaSalida(t, func() {
		runCheck([]string{"--lang", "python", "--dir", dir, "--format", "json",
			"--config", cfgPath})
	})
	if !strings.Contains(out, `"untested": []`) {
		t.Errorf("untested vacío debe serializarse como [], got:\n%s", out)
	}
	var rep map[string]any
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("JSON inválido: %v\n%s", err, out)
	}
}

func TestCheck_JSONConSkipped(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bin.py"), []byte("\x00\x01\x02BM"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	out := capturaSalida(t, func() {
		runCheck([]string{"--lang", "python", "--dir", dir, "--format", "json",
			"--config", cfgPath})
	})
	var rep struct {
		Skipped []struct {
			Path   string `json:"path"`
			Reason string `json:"reason"`
		} `json:"skipped"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("JSON inválido: %v\n%s", err, out)
	}
	if len(rep.Skipped) != 1 || rep.Skipped[0].Path != "bin.py" {
		t.Errorf("skipped = %+v, want [bin.py]", rep.Skipped)
	}
}

func TestCheck_JSONPackageMode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	cfgPath := creaConfigVacio(t, dir)

	out := capturaSalida(t, func() {
		runCheck([]string{"--lang", "go", "--dir", dir, "--format", "json",
			"--config", cfgPath})
	})
	var rep map[string]any
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("JSON inválido: %v\n%s", err, out)
	}
	if _, ok := rep["untested"]; !ok {
		t.Error("falta la clave untested")
	}
	if _, ok := rep["skipped"]; !ok {
		t.Error("falta la clave skipped")
	}
}
