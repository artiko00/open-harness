package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 8.9 / 8.15 — un paquete Go con 6 fuentes y un zzz_test.go sin marcador
// (solo 'package svc') queda sin cobertura: los 6 se reportan y exit 1.
func TestFase8_GoTestVacioNoCubre(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 6; i++ {
		os.WriteFile(filepath.Join(dir, "svc"+string(rune('a'+i))+".go"),
			[]byte("package svc"), 0644)
	}
	os.WriteFile(filepath.Join(dir, "zzz_test.go"), []byte("package svc"), 0644)

	cov, err := checkCoverage(config{language: "go", root: dir})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if cov.scanned != 6 {
		t.Errorf("scanned = %d, want 6", cov.scanned)
	}
	if len(cov.untested) != 1 {
		t.Fatalf("untested = %v, want el paquete sin test", cov.untested)
	}

	var code int
	capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "go", "--dir", dir, "--fail", "--no-color"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1 (paquete sin test real)", code)
	}
}

// 8.15 — modo por archivo: un *_test sin marcador no cubre a su fuente.
func TestFase8_PerFileTestVacioNoCubre(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "foo_test.py"), []byte("x = 1"), 0644)

	cov, err := checkCoverage(config{language: "python", root: dir})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if len(cov.untested) != 1 || cov.untested[0] != "foo.py" {
		t.Errorf("untested = %v, want [foo.py] (test vacío no cubre)", cov.untested)
	}
}

// 8.16 / 8.10 — test_calc.py (pytest, prefijo) no se reporta a sí mismo y SÍ
// cubre app/calc.py (mirror app<->tests, 8.18).
func TestFase8_PytestPrefijoNoEsFuenteYCubre(t *testing.T) {
	cov, err := checkCoverage(config{language: "python", root: "testdata/python_project"})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if cov.scanned != 1 {
		t.Errorf("scanned = %d, want 1 (test_calc.py no es fuente)", cov.scanned)
	}
	if len(cov.untested) != 0 {
		t.Errorf("untested = %v, want vacío (tests/test_calc.py cubre app/calc.py)", cov.untested)
	}
}

// 8.11 / 8.17 — __init__.py, settings.py, *_pb2.py y migraciones NO se reportan.
func TestFase8_ExclusionesPorDefecto(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "__init__.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "settings.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "foo_pb2.py"), []byte("x = 1"), 0644)
	os.MkdirAll(filepath.Join(dir, "migrations"), 0755)
	os.WriteFile(filepath.Join(dir, "migrations", "0001_init.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "real.py"), []byte("x = 1"), 0644)

	cov, err := checkCoverage(config{language: "python", root: dir})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if len(cov.untested) != 1 || cov.untested[0] != "real.py" {
		t.Errorf("untested = %v, want [real.py]", cov.untested)
	}
}

// 8.17 — la lista de exclusiones es configurable vía cfg.notest.
func TestFase8_ExclusionesConfigurables(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "skipme.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "real.py"), []byte("x = 1"), 0644)

	cov, err := checkCoverage(config{language: "python", root: dir, notest: []string{"skipme.py"}})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if len(cov.untested) != 1 || cov.untested[0] != "real.py" {
		t.Errorf("untested = %v, want [real.py] (skipme.py excluido por config)", cov.untested)
	}
}

// 8.17 — main.go y doc.go se excluyen por defecto en modo package.
func TestFase8_GoMainDocExcluidos(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "doc.go"), []byte("package main"), 0644)

	cov, err := checkCoverage(config{language: "go", root: dir})
	if err != nil {
		t.Fatalf("checkCoverage: %v", err)
	}
	if cov.scanned != 0 || len(cov.untested) != 0 {
		t.Errorf("main.go/doc.go deberían excluirse: scanned=%d untested=%v", cov.scanned, cov.untested)
	}
}

// 8.14 / 8.8 — un --lang no soportado (typo) va a STDERR con la lista y exit 1.
func TestFase8_LangNoSoportado(t *testing.T) {
	var code int
	errOut := capturaStderr(t, func() {
		code = runCheck([]string{"--lang", "typscript", "--dir", "testdata/ts_project"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "unsupported language") {
		t.Errorf("stderr debe indicar lenguaje no soportado: %q", errOut)
	}
	for _, l := range supportedLanguages() {
		if !strings.Contains(errOut, l) {
			t.Errorf("stderr debe listar %q: %q", l, errOut)
		}
	}
}

// contieneMarcador — rama de error al abrir un archivo inexistente.
func TestFase8_ContieneMarcadorArchivoInexistente(t *testing.T) {
	if contieneMarcador(filepath.Join(t.TempDir(), "no-existe.go"), []string{"func Test"}) {
		t.Error("un archivo inexistente no puede contener marcadores")
	}
}
