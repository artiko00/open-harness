package configload

import (
	"os"
	"path/filepath"
	"testing"
)

type testCfg struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("escribiendo %s: %v", name, err)
	}
}

// --- PackageJSON ---

func TestPackageJSONArchivoAusente(t *testing.T) {
	cfg, ok, err := PackageJSON[testCfg](t.TempDir(), "dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPackageJSONSeccionAusente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"otro":{"name":"x"}}`)
	cfg, ok, err := PackageJSON[testCfg](dir, "dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPackageJSONSeccionPresente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"dupelens":{"name":"a","enabled":true}}`)
	cfg, ok, err := PackageJSON[testCfg](dir, "dupelens")
	if err != nil || !ok || cfg.Name != "a" || !cfg.Enabled {
		t.Fatalf("esperado config poblada, got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPackageJSONArchivoInvalido(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{no-json`)
	_, ok, err := PackageJSON[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error, got ok=%v err=%v", ok, err)
	}
}

func TestPackageJSONSeccionInvalida(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"dupelens":{"enabled":"nope"}}`)
	_, ok, err := PackageJSON[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error, got ok=%v err=%v", ok, err)
	}
}

func TestPackageJSONErrorLectura(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "package.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, ok, err := PackageJSON[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error de lectura, got ok=%v err=%v", ok, err)
	}
}

// --- Composer ---

func TestComposerArchivoAusente(t *testing.T) {
	cfg, ok, err := Composer[testCfg](t.TempDir(), "dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestComposerSeccionAusente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"extra":{"open-harness":{"otro":{}}}}`)
	cfg, ok, err := Composer[testCfg](dir, "dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestComposerSeccionPresente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"extra":{"open-harness":{"dupelens":{"name":"b","enabled":true}}}}`)
	cfg, ok, err := Composer[testCfg](dir, "dupelens")
	if err != nil || !ok || cfg.Name != "b" {
		t.Fatalf("esperado config poblada, got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestComposerArchivoInvalido(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{no-json`)
	_, ok, err := Composer[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error, got ok=%v err=%v", ok, err)
	}
}

func TestComposerSeccionInvalida(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"extra":{"open-harness":{"dupelens":{"enabled":"nope"}}}}`)
	_, ok, err := Composer[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error, got ok=%v err=%v", ok, err)
	}
}

func TestComposerErrorLectura(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "composer.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, ok, err := Composer[testCfg](dir, "dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error de lectura, got ok=%v err=%v", ok, err)
	}
}

// --- Pyproject ---

func TestPyprojectArchivoAusente(t *testing.T) {
	cfg, ok, err := Pyproject[testCfg](t.TempDir(), "tool.dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPyprojectSeccionAusente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[tool.otro]\nname = \"x\"\n")
	cfg, ok, err := Pyproject[testCfg](dir, "tool.dupelens")
	if err != nil || ok || cfg != (testCfg{}) {
		t.Fatalf("esperado (zero,false,nil), got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPyprojectSeccionPresente(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[tool.dupelens]\nname = \"c\"\nenabled = true\n")
	cfg, ok, err := Pyproject[testCfg](dir, "tool.dupelens")
	if err != nil || !ok || cfg.Name != "c" || !cfg.Enabled {
		t.Fatalf("esperado config poblada, got (%v,%v,%v)", cfg, ok, err)
	}
}

func TestPyprojectTomlminError(t *testing.T) {
	dir := t.TempDir()
	// sintaxis fuera del subset soportado por tomlmin
	writeFile(t, dir, "pyproject.toml", "= sin_clave\n")
	_, ok, err := Pyproject[testCfg](dir, "tool.dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error tomlmin, got ok=%v err=%v", ok, err)
	}
}

func TestPyprojectSeccionInvalida(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[tool.dupelens]\nenabled = \"nope\"\n")
	_, ok, err := Pyproject[testCfg](dir, "tool.dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error de unmarshal, got ok=%v err=%v", ok, err)
	}
}

func TestPyprojectErrorLectura(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "pyproject.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, ok, err := Pyproject[testCfg](dir, "tool.dupelens")
	if err == nil || ok {
		t.Fatalf("esperado error de lectura, got ok=%v err=%v", ok, err)
	}
}
