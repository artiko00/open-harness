package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// captureStdout ejecuta fn redirigiendo stdout y devuelve lo impreso.
func captureStdout(t *testing.T, fn func()) string {
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
	var b bytes.Buffer
	io.Copy(&b, r)
	r.Close()
	return b.String()
}

// motivoDe busca el motivo registrado para un path relativo.
func motivoDe(skips []pathmatch.Skip, rel string) string {
	for _, s := range skips {
		if s.Path == rel {
			return s.Reason
		}
	}
	return ""
}

func cfgBase() Config {
	return Config{Default: DefaultConfig{MaxLines: 100}, Exclude: []string{}}
}

func TestScan_FIFONoCuelga(t *testing.T) {
	dir := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(dir, "pipe.go"), 0644); err != nil {
		t.Skip("mkfifo no soportado en este sistema")
	}
	writeFile(t, dir, "main.go", lines(10))

	results, skips, err := scan(dir, cfgBase(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := motivoDe(skips, "pipe.go"); got != pathmatch.ReasonNotRegular {
		t.Errorf("motivo del FIFO = %q, quiero %q", got, pathmatch.ReasonNotRegular)
	}
	if len(results) != 1 {
		t.Errorf("esperaba 1 archivo analizado, hay %d", len(results))
	}
}

func TestScan_ArchivoSinPermisoEsReadError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el chmod 0000 no bloquea la lectura")
	}
	dir := t.TempDir()
	writeFile(t, dir, "locked.go", lines(10))
	locked := filepath.Join(dir, "locked.go")
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	_, skips, err := scan(dir, cfgBase(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := motivoDe(skips, "locked.go"); got != pathmatch.ReasonReadError {
		t.Errorf("motivo = %q, quiero %q", got, pathmatch.ReasonReadError)
	}
}

func TestRunCheck_ReadErrorRompeElGate(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el chmod 0000 no bloquea la lectura")
	}
	dir := t.TempDir()
	writeFile(t, dir, "locked.go", lines(10))
	locked := filepath.Join(dir, "locked.go")
	os.Chmod(locked, 0000)
	defer os.Chmod(locked, 0644)

	var code int
	out := captureStdout(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, quiero 1", code)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener OK:\n%s", out)
	}
	if !strings.Contains(out, pathmatch.ReasonReadError) {
		t.Errorf("falta el motivo read error en la salida:\n%s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("falta el conteo de skipped en la salida:\n%s", out)
	}
}

func TestRunCheck_BinarioNoRompeElGate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "image.jpg", "no es realmente un jpg")
	writeFile(t, dir, "main.go", lines(10))

	var code int
	out := captureStdout(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit = %d, quiero 0", code)
	}
	if !strings.Contains(out, "image.jpg") || !strings.Contains(out, pathmatch.ReasonBinary) {
		t.Errorf("el binario debe figurar en SKIPPED:\n%s", out)
	}
	if !strings.Contains(out, "OK: all 1 files within limits (1 skipped)") {
		t.Errorf("linea final incorrecta:\n%s", out)
	}
}

func TestRunCheck_LineaMuyLargaRompeElGate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "huge.go", strings.Repeat("a", pathmatch.BufferLimit+1)+"\n")

	var code int
	out := captureStdout(t, func() {
		code = runCheck([]string{"--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, quiero 1", code)
	}
	if !strings.Contains(out, pathmatch.ReasonLineTooLong) {
		t.Errorf("falta el motivo de linea larga:\n%s", out)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener OK:\n%s", out)
	}
}

func TestReport_SkipsColor(t *testing.T) {
	skips := []pathmatch.Skip{{Path: "a.png", Reason: pathmatch.ReasonBinary}}
	out := captureStdout(t, func() {
		if n := reportConsole([]FileResult{}, skips, false); n != 0 {
			t.Errorf("violaciones = %d, quiero 0", n)
		}
	})
	if !strings.Contains(out, "SKIPPED") || !strings.Contains(out, "(1 skipped)") {
		t.Errorf("salida inesperada:\n%s", out)
	}
}

func TestReport_ViolacionesConSkips(t *testing.T) {
	results := []FileResult{{RelPath: "big.go", Lines: 150, MaxLines: 100}}
	skips := []pathmatch.Skip{{Path: "a.png", Reason: pathmatch.ReasonBinary}}
	out := captureStdout(t, func() {
		if n := reportConsole(results, skips, true); n != 1 {
			t.Errorf("violaciones = %d, quiero 1", n)
		}
	})
	if !strings.Contains(out, "SUMMARY: 1 violation(s) in 1 file(s) scanned, 1 skipped") {
		t.Errorf("linea final incorrecta:\n%s", out)
	}
}
