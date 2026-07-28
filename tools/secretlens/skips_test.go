package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// capturarSalida ejecuta fn redirigiendo stdout y devuelve lo impreso.
func capturarSalida(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig

	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	r.Close()
	return sb.String()
}

// buscarSkip devuelve el motivo registrado para una ruta, o "" si no está.
func buscarSkip(skips []pathmatch.Skip, path string) string {
	for _, s := range skips {
		if s.Path == path {
			return s.Reason
		}
	}
	return ""
}

func TestScan_FIFONoCuelga(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "pipe.go")
	if err := syscall.Mkfifo(fifo, 0644); err != nil {
		t.Skip("el sistema no soporta FIFOs:", err)
	}

	done := make(chan []pathmatch.Skip, 1)
	go func() {
		_, skips, _, _ := scan(dir, defaultConfig)
		done <- skips
	}()

	select {
	case skips := <-done:
		if got := buscarSkip(skips, "pipe.go"); got != pathmatch.ReasonNotRegular {
			t.Errorf("motivo = %q, want %q", got, pathmatch.ReasonNotRegular)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("scan se colgó con un FIFO")
	}
}

func TestScan_ArchivoSinPermisoEsReadError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el chmod 0000 no bloquea la lectura")
	}
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked.go")
	writeTestFile(t, dir, "locked.go", "package main\n")
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	_, skips, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if got := buscarSkip(skips, "locked.go"); got != pathmatch.ReasonReadError {
		t.Errorf("motivo = %q, want %q", got, pathmatch.ReasonReadError)
	}
	if !pathmatch.AnyFailsGate(skips) {
		t.Error("un read error debe romper el gate")
	}
}

func TestRunCheck_ReadErrorExit1SinOK(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el chmod 0000 no bloquea la lectura")
	}
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked.go")
	writeTestFile(t, dir, "locked.go", "package main\n")
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	var code int
	out := capturarSalida(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener OK:, got:\n%s", out)
	}
	if !strings.Contains(out, "SKIPPED (1 file(s) not analyzed):") {
		t.Errorf("falta encabezado SKIPPED, got:\n%s", out)
	}
	if !strings.Contains(out, pathmatch.ReasonReadError) {
		t.Errorf("falta el motivo read error, got:\n%s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("falta el conteo de omitidos, got:\n%s", out)
	}
}

func TestScan_BinarioEsSkipBinary(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "logo.png", "fake")
	writeTestFile(t, dir, "disguised.go", string([]byte{0x00, 'h', 'i'}))

	_, skips, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if got := buscarSkip(skips, "logo.png"); got != pathmatch.ReasonBinary {
		t.Errorf("logo.png motivo = %q, want %q", got, pathmatch.ReasonBinary)
	}
	if got := buscarSkip(skips, "disguised.go"); got != pathmatch.ReasonBinary {
		t.Errorf("disguised.go motivo = %q, want %q", got, pathmatch.ReasonBinary)
	}
	if pathmatch.AnyFailsGate(skips) {
		t.Error("un binario no debe romper el gate")
	}
}

func TestRunCheck_BinarioNoCambiaExitCode(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "logo.png", "fake")

	var code int
	out := capturarSalida(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "OK: no secrets detected (1 skipped)") {
		t.Errorf("esperaba OK con conteo de omitidos, got:\n%s", out)
	}
}

func TestScan_LineaMuyLargaEsSkip(t *testing.T) {
	dir := t.TempDir()
	linea := strings.Repeat("a", pathmatch.BufferLimit+1)
	writeTestFile(t, dir, "bundle.min.js", linea)

	_, skips, _, err := scan(dir, defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if got := buscarSkip(skips, "bundle.min.js"); got != pathmatch.ReasonLineTooLong {
		t.Errorf("motivo = %q, want %q", got, pathmatch.ReasonLineTooLong)
	}
}

func TestRunCheck_LineaMuyLargaExit1(t *testing.T) {
	dir := t.TempDir()
	linea := strings.Repeat("a", pathmatch.BufferLimit+1)
	writeTestFile(t, dir, "bundle.min.js", linea)

	var code int
	out := capturarSalida(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener OK:, got:\n%s", out)
	}
	if !strings.Contains(out, pathmatch.ReasonLineTooLong) {
		t.Errorf("falta el motivo de línea larga, got:\n%s", out)
	}
}

func TestReport_SkipsConHallazgos(t *testing.T) {
	findings := []Finding{
		{RelPath: "a.go", Line: 1, Content: "secret=abc", RuleName: "Generic", Severity: "high"},
	}
	skips := []pathmatch.Skip{{Path: "b.bin", Reason: pathmatch.ReasonBinary}}

	out := capturarSalida(t, func() { report(findings, skips, true) })
	if !strings.Contains(out, "SUMMARY: 1 potential secret(s) found — review before pushing, 1 skipped") {
		t.Errorf("resumen incorrecto, got:\n%s", out)
	}
}

func TestReport_SkipsColor(t *testing.T) {
	skips := []pathmatch.Skip{{Path: "b.bin", Reason: pathmatch.ReasonBinary}}
	out := capturarSalida(t, func() { report(nil, skips, false) })
	if !strings.Contains(out, "SKIPPED") {
		t.Errorf("falta SKIPPED en modo color, got:\n%s", out)
	}
	if !strings.Contains(out, "(1 skipped)") {
		t.Errorf("falta conteo en modo color, got:\n%s", out)
	}
}

func TestReport_SinSkipsNoImprimeSeccion(t *testing.T) {
	out := capturarSalida(t, func() { report(nil, nil, true) })
	if strings.Contains(out, "SKIPPED") {
		t.Errorf("no debía imprimir SKIPPED, got:\n%s", out)
	}
	if !strings.Contains(out, "OK: no secrets detected\n") {
		t.Errorf("esperaba OK limpio, got:\n%s", out)
	}
}

func TestScanFile_MotivoReadError(t *testing.T) {
	compiled, _ := compilePatterns(defaultPatterns())
	_, reason := scanFile("/nonexistent/file.go", "file.go", compiled, Config{})
	if reason != pathmatch.ReasonReadError {
		t.Errorf("motivo = %q, want %q", reason, pathmatch.ReasonReadError)
	}
}
