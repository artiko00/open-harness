package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ComposerFallback(t *testing.T) {
	dir := t.TempDir()
	j := `{"extra":{"open-harness":{"testlens":{"language":"php"}}}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(j), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "php" {
		t.Errorf("expected language=php, got %q", cfg.Language)
	}
}

func TestLoadConfig_ComposerNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"name":"x"}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language, got %q", cfg.Language)
	}
}

func TestLoadConfig_PackageJSONBeatsComposerJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"testlens":{"language":"php"}}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"testlens":{"language":"typescript"}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "typescript" {
		t.Errorf("expected package.json to win, got %q", cfg.Language)
	}
}

func TestLoadConfig_ComposerMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{invalid`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error for malformed composer.json")
	}
}

func TestLoadConfig_ComposerReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "composer.json"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error when composer.json is a directory")
	}
}

func TestLoadConfig_ComposerTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"testlens":42}}}`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "testlens.json")); err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestLoadConfig_ComposerAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"testlens":{"language":""}}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "testlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language applied, got %q", cfg.Language)
	}
}
