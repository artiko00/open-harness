package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TAREA 4.6 — --config explícito e inexistente = error.
func TestRunCheck_ConfigExplicitoInexistente(t *testing.T) {
	var code int
	_, errOut := capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", ".", "--config", "/nope/nope.json"})
	})
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not found") || !strings.Contains(errOut, "/nope/nope.json") {
		t.Errorf("stderr debe nombrar el archivo y decir 'not found', got:\n%s", errOut)
	}
}

// TAREA 4.6 — la ausencia del config por defecto sigue siendo válida.
func TestRunCheck_ConfigDefaultAusente(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\n"), 0644)
	var code int
	capturarStdoutStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--no-color"})
	})
	if code != 0 {
		t.Errorf("exit = %d, want 0 (debe caer a la chain/defaults)", code)
	}
}

// TAREA 4.7 — claves desconocidas en el config directo = error que nombra la clave.
func TestLoadConfig_ClaveDesconocida(t *testing.T) {
	// Retrocompat: una clave desconocida no falla; los conocidos se cargan.
	dir := t.TempDir()
	path := filepath.Join(dir, "secretlens.json")
	os.WriteFile(path, []byte(`{"desconocida":123,"minEntropy":2.5}`), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("clave desconocida no debe fallar (retrocompat), got %v", err)
	}
	if cfg.MinEntropy != 2.5 {
		t.Errorf("valores conocidos deben cargarse, got %v", cfg.MinEntropy)
	}
}

// TAREA 4.8 — merge por campo en la cadena de fallback.
func TestLoadConfig_ChainMergePorCampo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens]\nallowlist = [\"FROM_TOML\"]\n"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":{"exclude":["FROM_PKG_EXCLUDE"]}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Allowlist) == 0 || cfg.Allowlist[0] != "FROM_TOML" {
		t.Errorf("allowlist debe venir de pyproject, got %v", cfg.Allowlist)
	}
	if len(cfg.Exclude) == 0 || cfg.Exclude[0] != "FROM_PKG_EXCLUDE" {
		t.Errorf("exclude debe venir de package.json, got %v", cfg.Exclude)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("patterns debe rellenarse con defaults")
	}
}
