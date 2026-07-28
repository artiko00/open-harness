package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// captureRunCheck ejecuta runCheck redirigiendo stdout y stderr a pipes,
// y devuelve lo capturado junto al exit code.
func captureRunCheck(t *testing.T, args []string) (stdout, stderr string, code int) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	outC := make(chan string)
	errC := make(chan string)
	go func() { var b bytes.Buffer; b.ReadFrom(rOut); outC <- b.String() }()
	go func() { var b bytes.Buffer; b.ReadFrom(rErr); errC <- b.String() }()
	code = runCheck(args)
	wOut.Close()
	wErr.Close()
	stdout = <-outC
	stderr = <-errC
	os.Stdout, os.Stderr = origOut, origErr
	return
}

// lastLine devuelve la última línea no vacía del texto.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}

func TestContract_FormatJSON_valido_y_completo(t *testing.T) {
	stdout, stderr, code := captureRunCheck(t, []string{"--dir", "testdata/e2e_fixture", "--format", "json"})
	if code != 0 && code != 1 {
		t.Fatalf("code = %d, want 0 o 1", code)
	}
	if strings.Contains(stdout, "\033[") {
		t.Errorf("json no debe llevar códigos ANSI, stdout: %q", stdout)
	}
	if stderr != "" {
		t.Errorf("json: stderr debe estar vacío, got %q", stderr)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout no es JSON válido: %v\n%s", err, stdout)
	}
	if _, ok := parsed["scannedFiles"]; !ok {
		t.Error("falta clave scannedFiles (conteo de escaneados)")
	}
	if _, ok := parsed["matches"]; !ok {
		t.Error("falta clave matches (hallazgos)")
	}
	if _, ok := parsed["skipped"]; !ok {
		t.Error("falta clave skipped (omitidos)")
	}
}

func TestContract_FormatInvalido_xml(t *testing.T) {
	stdout, stderr, code := captureRunCheck(t, []string{"--dir", "testdata/e2e_fixture", "--format", "xml"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "console, json") {
		t.Errorf("stderr debe mencionar los formatos válidos, got %q", stderr)
	}
	if !strings.Contains(stderr, "xml") {
		t.Errorf("stderr debe mencionar el formato inválido, got %q", stderr)
	}
	if strings.Contains(stdout, "OK:") || strings.Contains(stdout, "SUMMARY:") || strings.Contains(stdout, "DUPLICATES") {
		t.Errorf("formato inválido no debe emitir reporte a stdout, got %q", stdout)
	}
}

func TestContract_DirNoAccesible_stderr(t *testing.T) {
	stdout, stderr, code := captureRunCheck(t, []string{"--dir", "/no/existe/xyz123"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "not accessible") {
		t.Errorf("stderr debe indicar directorio inaccesible, got %q", stderr)
	}
	if stdout != "" {
		t.Errorf("stdout debe estar vacío, got %q", stdout)
	}
}

func TestContract_FailOnInvalido(t *testing.T) {
	_, stderr, code := captureRunCheck(t, []string{"--dir", "testdata/e2e_fixture", "--fail-on", "bogus"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "exact, renamed, all") {
		t.Errorf("stderr debe listar los valores válidos de fail-on, got %q", stderr)
	}
}

func TestContract_FailOnRenamed_NoRompePorExact(t *testing.T) {
	// El par exacto de la fixture rompería --fail-on=exact pero NO --fail-on=renamed.
	_, _, code := captureRunCheck(t, []string{
		"--dir", "testdata/e2e_fixture", "--min-tokens", "30", "--fail", "--fail-on", "renamed"})
	if code != 0 {
		t.Errorf("un hallazgo exact no debe romper --fail-on=renamed, code = %d", code)
	}
}

func TestContract_FailOnAll_RompeConCualquiera(t *testing.T) {
	_, _, code := captureRunCheck(t, []string{
		"--dir", "testdata/e2e_fixture", "--min-tokens", "30", "--fail", "--fail-on", "all"})
	if code != 1 {
		t.Errorf("--fail-on=all debe romper con el par exacto, code = %d", code)
	}
}

func TestContract_MinTokensCero(t *testing.T) {
	_, stderr, code := captureRunCheck(t, []string{"--dir", "testdata/e2e_fixture", "--min-tokens", "0"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "min-tokens must be greater than 0") {
		t.Errorf("stderr inesperado: %q", stderr)
	}
}

func TestContract_MinTokensNegativo(t *testing.T) {
	_, stderr, code := captureRunCheck(t, []string{"--dir", "testdata/e2e_fixture", "--min-tokens", "-5"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "min-tokens must be greater than 0") {
		t.Errorf("stderr inesperado: %q", stderr)
	}
}

func TestContract_SinHallazgos_OK(t *testing.T) {
	stdout, _, code := captureRunCheck(t, []string{
		"--dir", "testdata/e2e_fixture", "--min-tokens", "5000", "--no-color"})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if got := lastLine(stdout); !strings.HasPrefix(got, "OK:") {
		t.Errorf("última línea debe empezar con OK:, got %q", got)
	}
}

func TestContract_ConHallazgos_SUMMARY(t *testing.T) {
	stdout, _, _ := captureRunCheck(t, []string{
		"--dir", "testdata/e2e_fixture", "--min-tokens", "30", "--no-color"})
	if !strings.Contains(stdout, "DUPLICATES (") {
		t.Errorf("debe existir encabezado DUPLICATES, got %q", stdout)
	}
	if got := lastLine(stdout); !strings.HasPrefix(got, "SUMMARY:") {
		t.Errorf("última línea debe empezar con SUMMARY:, got %q", got)
	}
}
