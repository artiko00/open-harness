package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

// capturaSalida ejecuta fn con stdout redirigido y devuelve lo impreso.
func capturaSalida(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	hecho := make(chan string, 1)
	go func() {
		var b strings.Builder
		io.Copy(&b, r)
		hecho <- b.String()
	}()
	fn()
	w.Close()
	os.Stdout = old
	return <-hecho
}

// creaFifo crea un FIFO o salta el test si el sistema no lo soporta.
func creaFifo(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0644); err != nil {
		t.Skipf("mkfifo no soportado: %v", err)
	}
}

// TAREA 2.10 — la detección de lenguaje debe honrar los excludes.

func TestDetectLanguageFromFiles_RespetaExclude(t *testing.T) {
	root := t.TempDir()
	nm := filepath.Join(root, "node_modules")
	os.MkdirAll(nm, 0755)
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(nm, fmt.Sprintf("m%d.rb", i)), []byte("x"), 0644)
	}
	src := filepath.Join(root, "src")
	os.MkdirAll(src, 0755)
	for i := 0; i < 6; i++ {
		os.WriteFile(filepath.Join(src, fmt.Sprintf("a%d.ts", i)), []byte("x"), 0644)
	}

	if got := detectLanguage(root, defaultConfig.Exclude); got != "typescript" {
		t.Fatalf("node_modules no debe decidir el lenguaje: got %q, want typescript", got)
	}
}

func TestDetectLanguageFromFiles_SinExcludeCuentaTodo(t *testing.T) {
	root := t.TempDir()
	nm := filepath.Join(root, "node_modules")
	os.MkdirAll(nm, 0755)
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(nm, fmt.Sprintf("m%d.rb", i)), []byte("x"), 0644)
	}
	if got := detectLanguage(root, nil); got != "ruby" {
		t.Fatalf("sin exclude debe detectar ruby, got %q", got)
	}
}

func TestDetectLanguageFromFiles_ErrorDeWalk(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	root := t.TempDir()
	sub := filepath.Join(root, "restricted")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "foo.go"), []byte("package foo"), 0644)
	os.Chmod(sub, 0000)
	defer os.Chmod(sub, 0755)

	if got := detectLanguage(root, nil); got != "" {
		t.Errorf("con walk abortado no hay lenguaje detectable, got %q", got)
	}
}

// Omisiones: FIFO, ilegible y binario.

func TestCheckCoverage_FifoOmitidoNoRegular(t *testing.T) {
	root := t.TempDir()
	creaFifo(t, filepath.Join(root, "pipe.py"))
	os.WriteFile(filepath.Join(root, "app.py"), []byte("x = 1"), 0644)

	violations, skips, err := checkCoverageCounts(config{language: "python", root: root, noColor: true})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if violations != 1 {
		t.Errorf("violations = %d, want 1 (solo app.py)", violations)
	}
	if len(skips) != 1 || skips[0].Reason != pathmatch.ReasonNotRegular {
		t.Fatalf("skips = %+v, want 1 con %q", skips, pathmatch.ReasonNotRegular)
	}
	if skips[0].Path != "pipe.py" {
		t.Errorf("skip path = %q, want pipe.py", skips[0].Path)
	}
}

func TestRunCheck_FifoEnSeccionSkipped(t *testing.T) {
	root := t.TempDir()
	creaFifo(t, filepath.Join(root, "pipe.py"))
	cfgPath := creaConfigVacio(t, root)

	var code int
	out := capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "python", "--dir", root, "--fail", "--no-color",
			"--config", cfgPath})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (not a regular file no rompe el gate)", code)
	}
	if !strings.Contains(out, "SKIPPED (1 file(s) not analyzed):") {
		t.Errorf("falta cabecera SKIPPED en:\n%s", out)
	}
	if !strings.Contains(out, "pipe.py    "+pathmatch.ReasonNotRegular) {
		t.Errorf("falta la línea del FIFO en:\n%s", out)
	}
}

func TestRunCheck_IlegibleRompeGate(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	root := t.TempDir()
	secreto := filepath.Join(root, "secreto.py")
	os.WriteFile(secreto, []byte("x = 1"), 0644)
	os.Chmod(secreto, 0000)
	defer os.Chmod(secreto, 0644)
	cfgPath := creaConfigVacio(t, root)

	var code int
	out := capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "python", "--dir", root, "--fail", "--no-color",
			"--config", cfgPath})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1 (read error rompe el gate)", code)
	}
	if !strings.Contains(out, "secreto.py    "+pathmatch.ReasonReadError) {
		t.Errorf("falta el skip por read error en:\n%s", out)
	}
	if strings.Contains(out, "OK:") {
		t.Errorf("la salida no debe contener OK:\n%s", out)
	}
}

func TestRunCheck_IlegibleSinFailNoRompe(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	root := t.TempDir()
	secreto := filepath.Join(root, "secreto.py")
	os.WriteFile(secreto, []byte("x = 1"), 0644)
	os.Chmod(secreto, 0000)
	defer os.Chmod(secreto, 0644)
	cfgPath := creaConfigVacio(t, root)

	var code int
	capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "python", "--dir", root, "--no-color",
			"--config", cfgPath})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 sin --fail", code)
	}
}

func TestRunCheck_BinarioNoRompeGate(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "bin.py"), []byte("\x00\x01\x02BM"), 0644)
	cfgPath := creaConfigVacio(t, root)

	var code int
	out := capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "python", "--dir", root, "--fail", "--no-color",
			"--config", cfgPath})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (binary no rompe el gate)", code)
	}
	if !strings.Contains(out, "bin.py    "+pathmatch.ReasonBinary) {
		t.Errorf("falta el skip por binario en:\n%s", out)
	}
}

// Modo package (Go): los mismos filtros aplican.

func TestCheckCoveragePackage_OmiteFifoYBinario(t *testing.T) {
	root := t.TempDir()
	creaFifo(t, filepath.Join(root, "pipe.go"))
	os.WriteFile(filepath.Join(root, "bin.go"), []byte("\x00\x01"), 0644)

	violations, skips, err := checkCoverageCounts(config{language: "go", root: root, noColor: true})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if violations != 0 {
		t.Errorf("violations = %d, want 0 (ningún .go analizable)", violations)
	}
	if len(skips) != 2 {
		t.Fatalf("skips = %+v, want 2", skips)
	}
	if pathmatch.AnyFailsGate(skips) {
		t.Error("ni FIFO ni binario deben romper el gate")
	}
}

func TestCheckCoveragePackage_WalkError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("chmod no bloquea como root")
	}
	root := t.TempDir()
	sub := filepath.Join(root, "restricted")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "foo.go"), []byte("package foo"), 0644)
	os.Chmod(sub, 0000)
	defer os.Chmod(sub, 0755)

	if _, _, err := checkCoverageCounts(config{language: "go", root: root}); err == nil {
		t.Error("modo package debe propagar el error de walk")
	}
}

// excludeEfectivo — config explícita gana sobre los defaults.

func TestExcludeEfectivo(t *testing.T) {
	if got := excludeEfectivo(config{}); len(got) != len(defaultConfig.Exclude) {
		t.Errorf("excludeEfectivo(vacío) = %v, want defaults", got)
	}
	if got := excludeEfectivo(config{exclude: []string{"solo"}}); len(got) != 1 || got[0] != "solo" {
		t.Errorf("excludeEfectivo(explícito) = %v, want [solo]", got)
	}
}
