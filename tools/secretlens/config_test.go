package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NotExists(t *testing.T) {
	cfg, err := loadConfig("/nonexistent/secretlens.json")
	if err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("expected default patterns")
	}
}

func TestLoadConfig_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secretlens.json")
	content := `{"patterns":[{"name":"test","pattern":"TEST_\\d+","severity":"high"}],"allowlist":["example"],"exclude":["node_modules"]}`
	os.WriteFile(path, []byte(content), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	// Patterns aditivos: defaults + el custom.
	if len(cfg.Patterns) != len(defaultPatterns())+1 {
		t.Errorf("expected defaults+1 patterns, got %d", len(cfg.Patterns))
	}
	if cfg.Patterns[len(cfg.Patterns)-1].Name != "test" {
		t.Errorf("expected custom pattern last, got %q", cfg.Patterns[len(cfg.Patterns)-1].Name)
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
	path := filepath.Join(dir, "secretlens.json")
	// empty patterns/allowlist/exclude → defaults applied
	os.WriteFile(path, []byte(`{"patterns":[],"allowlist":[],"exclude":[]}`), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("expected default patterns applied")
	}
	if len(cfg.Allowlist) == 0 {
		t.Error("expected default allowlist applied")
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
