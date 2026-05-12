package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NotExists(t *testing.T) {
	cfg, err := loadConfig("/nonexistent/testlens.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "auto" {
		t.Errorf("expected default language auto, got %q", cfg.Language)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default exclude list applied")
	}
}

func TestLoadConfig_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testlens.json")
	os.WriteFile(path, []byte(`{"language":"go","exclude":["node_modules","custom"]}`), 0644)

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Language != "go" {
		t.Errorf("expected go, got %q", cfg.Language)
	}
	if len(cfg.Exclude) != 2 || cfg.Exclude[1] != "custom" {
		t.Errorf("expected custom exclude, got %v", cfg.Exclude)
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
	path := filepath.Join(dir, "testlens.json")
	os.WriteFile(path, []byte(`{"language":"","exclude":[]}`), 0644)

	cfg, err := loadConfig(path)
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

func TestLoadConfig_ReadError(t *testing.T) {
	dir := t.TempDir()
	_, err := loadConfig(dir)
	if err == nil {
		t.Error("expected error reading directory as config file")
	}
}
