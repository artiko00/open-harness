package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NotExists(t *testing.T) {
	cfg, err := loadConfig("/nonexistent/linelens.json")
	if err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected default 100, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "linelens.json")
	os.WriteFile(path, []byte(`{"default":{"maxLines":200},"rules":[],"exclude":["node_modules"]}`), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Default.MaxLines != 200 {
		t.Errorf("expected 200, got %d", cfg.Default.MaxLines)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid}"), 0644)
	_, err := loadConfig(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "linelens.json")
	os.WriteFile(path, []byte(`{"default":{"maxLines":0},"exclude":[]}`), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Default.MaxLines != 100 {
		t.Errorf("expected default 100 applied, got %d", cfg.Default.MaxLines)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default excludes applied")
	}
}

func TestLoadConfig_ReadError(t *testing.T) {
	dir := t.TempDir()
	_, err := loadConfig(dir)
	if err == nil {
		t.Error("expected error reading directory as config file")
	}
}

func TestDefaultConfigJSON(t *testing.T) {
	s := defaultConfigJSON()
	if len(s) == 0 {
		t.Error("expected non-empty config JSON")
	}
}
