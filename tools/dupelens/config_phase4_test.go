package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 4.6 — --config explícito e inexistente = error a stderr y exit 1.
func TestRunCheck_explicitConfigMissing_error(t *testing.T) {
	_, stderr, code := captureRunCheck(t, []string{
		"--dir", "testdata/e2e_fixture", "--config", "/nope/nope.json"})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, `config file "/nope/nope.json" not found`) {
		t.Errorf("stderr debe indicar config no encontrado, got %q", stderr)
	}
}

// 4.6 — ausencia del config por DEFECTO sigue siendo válida (cae a la chain), exit 0.
func TestRunCheck_defaultConfigMissing_ok(t *testing.T) {
	dir := t.TempDir()
	scan := filepath.Join(dir, "src")
	if err := os.Mkdir(scan, 0o755); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	_, _, code := captureRunCheck(t, []string{"--dir", "src"})
	if code != 0 {
		t.Fatalf("code = %d, want 0 (ausencia de config por defecto es válida)", code)
	}
}

// 4.7 — clave desconocida en el config directo = error que la nombra.
func TestLoadConfig_unknownKey_warning(t *testing.T) {
	// Retrocompat: una clave desconocida no falla; los conocidos se cargan.
	dir := t.TempDir()
	writeTempFile(t, dir, "dupelens.json", `{"default":{"minTokens":77},"bogus":true}`)
	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("clave desconocida no debe fallar (retrocompat), got %v", err)
	}
	if cfg.Default.MinTokens != 77 {
		t.Errorf("valores conocidos deben cargarse, got %d", cfg.Default.MinTokens)
	}
}

// 4.x — un archivo malformado en la cadena de fallback propaga el error.
func TestLoadConfig_chainLoaderError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid`), 0o644)
	if _, err := loadConfig(filepath.Join(dir, "dupelens.json")); err == nil {
		t.Fatal("expected error propagado desde la cadena de fallback")
	}
}

// 4.8 — merge por campo en la cadena de fallback: pyproject aporta minTokens,
// package.json aporta exclude; ambos se aplican.
func TestLoadConfig_mergeChain_fieldLevel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.dupelens]\ndefault = { minTokens = 111 }\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dupelens":{"exclude":["only-pkg"]}}`), 0o644)

	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 111 {
		t.Errorf("minTokens debe venir de pyproject (111), got %d", cfg.Default.MinTokens)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "only-pkg" {
		t.Errorf("exclude debe venir de package.json, got %v", cfg.Exclude)
	}
	if cfg.Default.MinLines != 5 {
		t.Errorf("minLines debe heredar default=5, got %d", cfg.Default.MinLines)
	}
}
