package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PyprojectFallback(t *testing.T) {
	dir := t.TempDir()
	toml := `
[tool.secretlens]
allowlist = ["MY_PLACEHOLDER", "fixture-token"]
`
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(toml), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, a := range cfg.Allowlist {
		if a == "MY_PLACEHOLDER" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected allowlist from pyproject.toml, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_PyprojectNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[project]\nname = \"x\"\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("expected default patterns when [tool.secretlens] absent")
	}
}

func TestLoadConfig_DedicatedFileBeatsPyproject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens]\nallowlist = [\"FROM_TOML\"]\n"), 0644)
	os.WriteFile(filepath.Join(dir, "secretlens.json"),
		[]byte(`{"allowlist":["FROM_FILE"]}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Allowlist[0] != "FROM_FILE" {
		t.Errorf("expected dedicated file to win, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_PyprojectBeatsPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":{"allowlist":["FROM_PKG"]}}`), 0644)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens]\nallowlist = [\"FROM_TOML\"]\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Allowlist[0] != "FROM_TOML" {
		t.Errorf("expected pyproject to win over package.json, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_PyprojectMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens\nbroken"), 0644)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error for malformed pyproject.toml")
	}
}

func TestLoadConfig_PyprojectReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "pyproject.toml"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error when pyproject.toml is a directory")
	}
}

func TestLoadConfig_PyprojectTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens]\nallowlist = \"not-an-array\"\n"), 0644)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestLoadConfig_PyprojectAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.secretlens]\nallowlist = []\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) == 0 || len(cfg.Allowlist) == 0 || len(cfg.Exclude) == 0 {
		t.Error("expected defaults applied when pyproject omits them")
	}
}
