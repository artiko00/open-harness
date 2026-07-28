package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TAREA 4.6 — --config explicito e inexistente = error.
func TestCheck_ConfigExplicitoInexistente(t *testing.T) {
	var code int
	errOut := capturaStderr(t, func() {
		code = runCheck([]string{"--config", "/nope/nope.json", "--dir", t.TempDir()})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, `config file "/nope/nope.json" not found`) {
		t.Errorf("stderr debe indicar config no encontrado: %q", errOut)
	}
}

// TAREA 4.6 — la ausencia del config por DEFECTO sigue siendo valida.
func TestCheck_ConfigDefaultAusenteValido(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.py"), []byte("x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "foo_test.py"), []byte("x = 1"), 0644)

	var code int
	capturaSalida(t, func() {
		code = runCheck([]string{"--lang", "python", "--dir", dir, "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (config por defecto ausente cae a defaults)", code)
	}
}

// TAREA 4.7 — clave desconocida = error nombrando la clave.
func TestCheck_ConfigClaveDesconocida(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testlens.json")
	os.WriteFile(path, []byte(`{"language":"go","cosaRara":1}`), 0644)

	var code int
	errOut := capturaStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", path})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (aviso, no error)", code)
	}
	if !strings.Contains(errOut, "cosaRara") || !strings.Contains(errOut, "warning") {
		t.Errorf("stderr debe avisar y nombrar la clave desconocida: %q", errOut)
	}
}

// TAREA 4.3 — clave legada "skip" => error nombrando la clave + sugerencia "exclude".
func TestCheck_ConfigClaveSkipSugiereExclude(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testlens.json")
	os.WriteFile(path, []byte(`{"skip":["node_modules"]}`), 0644)

	var code int
	errOut := capturaStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", path})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (aviso, no error)", code)
	}
	if !strings.Contains(errOut, `"skip"`) {
		t.Errorf("stderr debe nombrar la clave skip: %q", errOut)
	}
	if !strings.Contains(errOut, "exclude") {
		t.Errorf("stderr debe sugerir usar exclude: %q", errOut)
	}
}

// TAREA 4.8 — merge por campo: pyproject aporta language, package.json aporta exclude.
func TestLoadConfig_MergeFallbackChain(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens]\nlanguage = \"go\"\n"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":{"exclude":["custom"]}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "go" {
		t.Errorf("language = %q, want go (de pyproject)", cfg.Language)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "custom" {
		t.Errorf("exclude = %v, want [custom] (de package.json)", cfg.Exclude)
	}
}
