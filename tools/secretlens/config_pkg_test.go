package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PackageJSONFallback(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"x","secretlens":{"allowlist":["MY_PLACEHOLDER"]}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, a := range cfg.Allowlist {
		if a == "MY_PLACEHOLDER" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected allowlist from package.json, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_PackageJSONNoKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) == 0 {
		t.Error("expected default patterns when no secretlens key")
	}
}

func TestLoadConfig_DedicatedFileWinsOverPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":{"allowlist":["FROM_PKG"]}}`), 0644)
	os.WriteFile(filepath.Join(dir, "secretlens.json"),
		[]byte(`{"allowlist":["FROM_FILE"]}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Allowlist[0] != "FROM_FILE" {
		t.Errorf("expected dedicated file allowlist to win, got %v", cfg.Allowlist)
	}
}

func TestLoadConfig_PackageJSONInvalidKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":"not-an-object"}`), 0644)

	_, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err == nil {
		t.Error("expected error for malformed secretlens key in package.json")
	}
}

func TestLoadConfig_PackageJSONMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid`), 0644)

	_, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err == nil {
		t.Error("expected error for malformed package.json")
	}
}

func TestLoadConfig_PackageJSONReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "package.json"), 0755)
	_, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err == nil {
		t.Error("expected error when package.json is a directory")
	}
}

func TestLoadConfig_PackageJSONAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"secretlens":{"patterns":[],"allowlist":[],"exclude":[]}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "secretlens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
