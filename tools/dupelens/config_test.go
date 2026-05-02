package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_missingFileReturnsDefaults(t *testing.T) {
	cfg, err := loadConfig(filepath.Join(t.TempDir(), "no-such.json"))
	if err != nil {
		t.Fatalf("missing config no debe ser error, got: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("expected default minTokens=50, got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_malformedJSON_includesPathAndHint(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "broken.json")
	writeTempFile(t, dir, "broken.json", "{ not valid json")
	_, err := loadConfig(bad)
	if err == nil {
		t.Fatal("expected error con malformed JSON, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, bad) && !strings.Contains(msg, "broken.json") {
		t.Errorf("error debe mencionar el path del config, got: %q", msg)
	}
	if !strings.Contains(strings.ToLower(msg), "config") {
		t.Errorf("error debe mencionar 'config' para que el usuario sepa qué falló, got: %q", msg)
	}
}

func TestLoadConfig_validJSONOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "ok.json")
	writeTempFile(t, dir, "ok.json", `{"default":{"minTokens":99,"minLines":3}}`)
	cfg, err := loadConfig(valid)
	if err != nil {
		t.Fatalf("valid config no debe fallar: %v", err)
	}
	if cfg.Default.MinTokens != 99 {
		t.Errorf("minTokens override no aplicado, got %d", cfg.Default.MinTokens)
	}
	if cfg.Default.MinLines != 3 {
		t.Errorf("minLines override no aplicado, got %d", cfg.Default.MinLines)
	}
}

func TestLoadConfig_partialOverridesFillFromDefaults(t *testing.T) {
	dir := t.TempDir()
	partial := filepath.Join(dir, "partial.json")
	writeTempFile(t, dir, "partial.json", `{"default":{"minTokens":77}}`)
	cfg, err := loadConfig(partial)
	if err != nil {
		t.Fatalf("partial config no debe fallar: %v", err)
	}
	if cfg.Default.MinTokens != 77 {
		t.Errorf("minTokens override = %d; want 77", cfg.Default.MinTokens)
	}
	if cfg.Default.MinLines != 5 {
		t.Errorf("minLines debe heredar default=5, got %d", cfg.Default.MinLines)
	}
	if len(cfg.Exclude) == 0 {
		t.Errorf("Exclude vacío debe heredar defaults, got 0")
	}
}
