package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PackageJSONFallback(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"x","linelens":{"default":{"maxLines":250}}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 250 {
		t.Errorf("expected maxLines from package.json (250), got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_PackageJSONNoKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected defaults when no linelens key, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_DedicatedFileWinsOverPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":{"default":{"maxLines":250}}}`), 0644)
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

func TestLoadConfig_PackageJSONInvalidKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":"not-an-object"}`), 0644)

	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error for malformed linelens key in package.json")
	}
}

func TestLoadConfig_PackageJSONMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid`), 0644)

	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error for malformed package.json")
	}
}

func TestLoadConfig_PackageJSONReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "package.json"), 0755)
	_, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err == nil {
		t.Error("expected error when package.json is a directory")
	}
}

func TestLoadConfig_PackageJSONAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":{"default":{"maxLines":0}}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected default 100 applied, got %d", cfg.Default.MaxLines)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default excludes applied")
	}
}
