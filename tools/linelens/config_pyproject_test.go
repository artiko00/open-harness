package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PyprojectFallback(t *testing.T) {
	dir := t.TempDir()
	toml := `
[tool.linelens]
default = { maxLines = 250 }
exclude = ["node_modules", "custom"]
`
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(toml), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 250 {
		t.Errorf("expected maxLines=250 from pyproject.toml, got %d", cfg.Default.MaxLines)
	}
	if len(cfg.Exclude) != 2 || cfg.Exclude[1] != "custom" {
		t.Errorf("expected custom exclude from pyproject.toml, got %v", cfg.Exclude)
	}
}

func TestLoadConfig_PyprojectNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[project]\nname = \"x\"\n"), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected defaults when no [tool.linelens], got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_DedicatedFileBeatsPyproject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.linelens]\ndefault = { maxLines = 250 }\n"), 0644)
	os.WriteFile(filepath.Join(dir, "linelens.json"),
		[]byte(`{"default":{"maxLines":80}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 80 {
		t.Errorf("expected dedicated file to win (80), got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_PyprojectBeatsPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":{"default":{"maxLines":300}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.linelens]\ndefault = { maxLines = 175 }\n"), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 175 {
		t.Errorf("expected pyproject (175) to win over package.json, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_PyprojectMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.linelens\nbroken"), 0644)
	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error for malformed pyproject.toml")
	}
}

func TestLoadConfig_PyprojectReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "pyproject.toml"), 0755)
	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error when pyproject.toml is a directory")
	}
}

func TestLoadConfig_PyprojectTypeMismatch(t *testing.T) {
	// TOML is syntactically valid, but `default` is a string not a table:
	// json.Unmarshal into Config.Default (a struct) fails.
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte(`[tool.linelens]`+"\ndefault = \"not-a-table\"\n"), 0644)
	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error for type mismatch in pyproject.toml [tool.linelens]")
	}
}

func TestLoadConfig_PyprojectAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.linelens]\ndefault = { maxLines = 0 }\n"), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected default 100 applied, got %d", cfg.Default.MaxLines)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default excludes applied when pyproject.toml omits them")
	}
}
