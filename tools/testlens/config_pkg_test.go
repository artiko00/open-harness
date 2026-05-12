package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PackageJSONFallback(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"x","testlens":{"language":"typescript","exclude":["custom"]}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "typescript" {
		t.Errorf("expected language from package.json (typescript), got %q", cfg.Language)
	}
	if len(cfg.Exclude) == 0 || cfg.Exclude[0] != "custom" {
		t.Errorf("expected exclude from package.json, got %v", cfg.Exclude)
	}
}

func TestLoadConfig_PackageJSONNoKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language when no testlens key, got %q", cfg.Language)
	}
}

func TestLoadConfig_DedicatedFileWinsOverPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":{"language":"typescript"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "testlens.json"),
		[]byte(`{"language":"go"}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "go" {
		t.Errorf("expected dedicated file to win (go), got %q", cfg.Language)
	}
}

func TestLoadConfig_PackageJSONInvalidKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":"not-an-object"}`), 0644)

	_, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err == nil {
		t.Error("expected error for malformed testlens key in package.json")
	}
}

func TestLoadConfig_PackageJSONMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid`), 0644)

	_, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err == nil {
		t.Error("expected error for malformed package.json")
	}
}

func TestLoadConfig_PackageJSONReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "package.json"), 0755)
	_, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err == nil {
		t.Error("expected error when package.json is a directory")
	}
}

func TestLoadConfig_PackageJSONAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":{"language":"","exclude":[]}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language auto applied, got %q", cfg.Language)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default excludes applied")
	}
}
