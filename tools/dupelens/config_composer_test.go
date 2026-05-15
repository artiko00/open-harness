package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ComposerFallback(t *testing.T) {
	dir := t.TempDir()
	j := `{"extra":{"open-harness":{"dupelens":{"default":{"minTokens":80}}}}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(j), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 80 {
		t.Errorf("expected minTokens=80 from composer.json, got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_ComposerNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"name":"x"}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("expected defaults, got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_PackageJSONBeatsComposerJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"dupelens":{"default":{"minTokens":80}}}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dupelens":{"default":{"minTokens":120}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 120 {
		t.Errorf("expected package.json (120) to win over composer.json, got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_ComposerMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{invalid`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "dupelens.json")); err == nil {
		t.Error("expected error for malformed composer.json")
	}
}

func TestLoadConfig_ComposerReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "composer.json"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "dupelens.json")); err == nil {
		t.Error("expected error when composer.json is a directory")
	}
}

func TestLoadConfig_ComposerTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"dupelens":"not-a-table"}}}`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "dupelens.json")); err == nil {
		t.Error("expected error for type mismatch in composer.json")
	}
}

func TestLoadConfig_ComposerAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"dupelens":{"default":{"minTokens":0}}}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("expected default 50, got %d", cfg.Default.MinTokens)
	}
}
