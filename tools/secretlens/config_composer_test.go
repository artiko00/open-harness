package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ComposerFallback(t *testing.T) {
	dir := t.TempDir()
	j := `{"extra":{"open-harness":{"secretlens":{"allowlist":["FROM_COMPOSER"]}}}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(j), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Allowlist[0] != "FROM_COMPOSER" {
		t.Errorf("expected allowlist from composer.json, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_ComposerNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"name":"x"}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("expected default patterns")
	}
}

func TestLoadConfig_PackageJSONBeatsComposerJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"secretlens":{"allowlist":["FROM_COMPOSER"]}}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":{"allowlist":["FROM_PKG"]}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Allowlist[0] != "FROM_PKG" {
		t.Errorf("expected package.json to win, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_ComposerMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{invalid`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error for malformed composer.json")
	}
}

func TestLoadConfig_ComposerReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "composer.json"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error when composer.json is a directory")
	}
}

func TestLoadConfig_ComposerTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"secretlens":"not-a-table"}}}`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "secretlens.json")); err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestLoadConfig_ComposerAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"secretlens":{"allowlist":[]}}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) == 0 || len(cfg.Allowlist) == 0 {
		t.Error("expected defaults applied")
	}
}
