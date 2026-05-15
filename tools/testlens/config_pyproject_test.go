package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PyprojectFallback(t *testing.T) {
	dir := t.TempDir()
	toml := `
[tool.testlens]
language = "typescript"
exclude = ["custom-skip"]
`
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(toml), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "typescript" {
		t.Errorf("expected language=typescript, got %q", cfg.Language)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "custom-skip" {
		t.Errorf("expected custom exclude, got %v", cfg.Exclude)
	}
}

func TestLoadConfig_PyprojectNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[project]\nname = \"x\"\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language, got %q", cfg.Language)
	}
}

func TestLoadConfig_DedicatedFileBeatsPyproject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens]\nlanguage = \"typescript\"\n"), 0644)
	os.WriteFile(filepath.Join(dir, "testlens.json"),
		[]byte(`{"language":"go"}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "go" {
		t.Errorf("expected dedicated file to win, got %q", cfg.Language)
	}
}

func TestLoadConfig_PyprojectBeatsPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":{"language":"go"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens]\nlanguage = \"python\"\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "python" {
		t.Errorf("expected pyproject (python) to win, got %q", cfg.Language)
	}
}

func TestLoadConfig_PyprojectMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens\nbroken"), 0644)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error for malformed pyproject.toml")
	}
}

func TestLoadConfig_PyprojectReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "pyproject.toml"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error when pyproject.toml is a directory")
	}
}

func TestLoadConfig_PyprojectTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens]\nlanguage = 123\n"), 0644)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error for type mismatch (number instead of string)")
	}
}

func TestLoadConfig_PyprojectAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.testlens]\nlanguage = \"\"\nexclude = []\n"), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" || len(cfg.Exclude) == 0 {
		t.Errorf("expected defaults applied: %v", cfg)
	}
}
