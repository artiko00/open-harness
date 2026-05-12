package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_PackageJSONFallback(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"x","dupelens":{"default":{"minTokens":80}}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 80 {
		t.Errorf("expected minTokens from package.json (80), got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_PackageJSONNoKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("expected defaults when no dupelens key, got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_DedicatedFileWinsOverPackageJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dupelens":{"default":{"minTokens":80}}}`), 0644)
	os.WriteFile(filepath.Join(dir, "dupelens.json"),
		[]byte(`{"default":{"minTokens":30}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 30 {
		t.Errorf("expected dedicated file to win (30), got %d", cfg.Default.MinTokens)
	}
}

func TestLoadConfig_PackageJSONInvalidKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dupelens":"not-an-object"}`), 0644)

	_, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err == nil {
		t.Error("expected error for malformed dupelens key in package.json")
	}
}

func TestLoadConfig_PackageJSONMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{invalid`), 0644)

	_, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err == nil {
		t.Error("expected error for malformed package.json")
	}
}

func TestLoadConfig_PackageJSONReadError(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "package.json"), 0755)
	_, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err == nil {
		t.Error("expected error when package.json is a directory")
	}
}

func TestLoadConfig_PackageJSONAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dupelens":{"default":{"minTokens":0,"minLines":0}}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "dupelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MinTokens != 50 {
		t.Errorf("expected default minTokens 50 applied, got %d", cfg.Default.MinTokens)
	}
	if cfg.Default.MinLines != 5 {
		t.Errorf("expected default minLines 5 applied, got %d", cfg.Default.MinLines)
	}
	if len(cfg.Exclude) == 0 {
		t.Error("expected default excludes applied")
	}
}
