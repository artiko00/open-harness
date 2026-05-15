package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ComposerFallback(t *testing.T) {
	dir := t.TempDir()
	j := `{"extra":{"open-harness":{"linelens":{"default":{"maxLines":175}}}}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(j), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 175 {
		t.Errorf("expected maxLines=175 from composer.json, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_ComposerNoSection(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"name":"x"}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected defaults, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_PackageJSONBeatsComposerJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"linelens":{"default":{"maxLines":175}}}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":{"default":{"maxLines":200}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 200 {
		t.Errorf("expected package.json (200) to win over composer.json, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_ComposerMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{invalid`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "linelens.json")); err == nil {
		t.Error("expected error for malformed composer.json")
	}
}

func TestLoadConfig_ComposerReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "composer.json"), 0755)
	if _, err := loadConfig(filepath.Join(dir, "linelens.json")); err == nil {
		t.Error("expected error when composer.json is a directory")
	}
}

func TestLoadConfig_ComposerTypeMismatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"linelens":"not-a-table"}}}`), 0644)
	if _, err := loadConfig(filepath.Join(dir, "linelens.json")); err == nil {
		t.Error("expected error for type mismatch in composer.json")
	}
}

func TestLoadConfig_ComposerAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"),
		[]byte(`{"extra":{"open-harness":{"linelens":{"default":{"maxLines":0}}}}}`), 0644)
	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected default 100, got %d", cfg.Default.MaxLines)
	}
}
