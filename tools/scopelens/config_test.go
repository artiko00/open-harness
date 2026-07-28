package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("escribiendo %s: %v", name, err)
	}
}

func TestLoadConfigJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":20,"base":"develop","excludeTests":true,"exclude":["docs/**"]}`)
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxFiles != 20 || cfg.Base != "develop" || !cfg.ExcludeTests || len(cfg.Exclude) != 1 {
		t.Fatalf("config mal cargada: %+v", cfg)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxFiles != defaultMaxFiles {
		t.Fatalf("maxFiles default = %d", cfg.MaxFiles)
	}
	if len(cfg.Exclude) != len(defaultExclude) {
		t.Fatalf("exclude default no aplicado: %d", len(cfg.Exclude))
	}
}

func TestLoadConfigMalformado(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":15,}`)
	_, err := loadConfig(dir)
	if err == nil || !strings.Contains(err.Error(), "scopelens.json") {
		t.Fatalf("esperaba error nombrando scopelens.json, got %v", err)
	}
}

func TestLoadConfigMaxFilesNegativo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":-3}`)
	_, err := loadConfig(dir)
	if err == nil || !strings.Contains(err.Error(), "maxFiles") {
		t.Fatalf("esperaba error nombrando maxFiles, got %v", err)
	}
}

func TestLoadConfigClaveDesconocida(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":15,"raro":1}`)
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatalf("clave desconocida no debe fallar: %v", err)
	}
	if cfg.MaxFiles != 15 {
		t.Fatalf("config ignorada: %+v", cfg)
	}
}

func TestLoadConfigErrorLectura(t *testing.T) {
	dir := t.TempDir()
	// scopelens.json es un directorio: ReadFile falla con un error != NotExist.
	if err := os.Mkdir(filepath.Join(dir, "scopelens.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(dir); err == nil {
		t.Fatal("esperaba error de lectura")
	}
}

func TestLoadConfigChain(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[tool.scopelens]\nmaxFiles = 10\n")
	writeFile(t, dir, "package.json", `{"scopelens":{"maxFiles":20,"base":"main"}}`)
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	// pyproject gana en maxFiles; package.json aporta base (pyproject no lo define).
	if cfg.MaxFiles != 10 || cfg.Base != "main" {
		t.Fatalf("merge por campo incorrecto: %+v", cfg)
	}
}

func TestLoadConfigChainComposer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"extra":{"open-harness":{"scopelens":{"maxFiles":8,"excludeTests":true}}}}`)
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxFiles != 8 || !cfg.ExcludeTests {
		t.Fatalf("composer mal cargado: %+v", cfg)
	}
}

func TestLoadConfigChainError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"scopelens": not json}`)
	if _, err := loadConfig(dir); err == nil {
		t.Fatal("esperaba error de package.json inválido")
	}
}
