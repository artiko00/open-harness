package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// capturaStderr ejecuta fn redirigiendo os.Stderr y devuelve lo escrito.
func capturaStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// contieneExclude indica si v está en xs.
func contieneExclude(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// TAREA 4.6: --config explícito e inexistente => error a stderr y exit 1.
func TestRunCheck_ExplicitConfigMissing(t *testing.T) {
	dir := t.TempDir()
	var code int
	out := capturaStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", "/nope/nope.json"})
	})
	if code != 1 {
		t.Fatalf("esperado exit 1, got %d", code)
	}
	if !strings.Contains(out, `config file "/nope/nope.json" not found`) {
		t.Errorf("stderr debe nombrar el archivo faltante, got %q", out)
	}
}

// TAREA 4.6: ausencia del config por DEFECTO sigue siendo válida (cae a la chain).
func TestRunCheck_DefaultConfigMissingIsValid(t *testing.T) {
	work := t.TempDir()
	scan := t.TempDir()
	os.WriteFile(filepath.Join(scan, "ok.go"), []byte("package main\n"), 0644)
	orig, _ := os.Getwd()
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	if code := runCheck([]string{"--dir", scan, "--no-color"}); code != 0 {
		t.Errorf("ausencia de config por defecto debe ser válida (exit 0), got %d", code)
	}
}

// TAREA 4.7 / 4.4: clave desconocida en el config directo => error nombrando la clave.
func TestLoadConfig_UnknownKey(t *testing.T) {
	// Retrocompat: una clave desconocida NO falla; los valores conocidos se cargan.
	dir := t.TempDir()
	path := filepath.Join(dir, "linelens.json")
	os.WriteFile(path, []byte(`{"excludes":["x"],"default":{"maxLines":42}}`), 0644)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("clave desconocida no debe fallar (retrocompat), got %v", err)
	}
	if cfg.Default.MaxLines != 42 {
		t.Errorf("los valores conocidos deben cargarse igual, got %d", cfg.Default.MaxLines)
	}
}

func TestRunCheck_UnknownKeyWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "linelens.json")
	os.WriteFile(path, []byte(`{"excludes":["x"]}`), 0644)
	var code int
	out := capturaStderr(t, func() {
		code = runCheck([]string{"--dir", dir, "--config", path})
	})
	if code != 0 {
		t.Fatalf("clave desconocida debe avisar y seguir (exit 0), got %d", code)
	}
	if !strings.Contains(out, "excludes") || !strings.Contains(out, "warning") {
		t.Errorf("stderr debe avisar y nombrar la clave, got %q", out)
	}
}

// TAREA 4.5: merge por campo en la cadena de fallback.
// pyproject aporta maxLines, package.json aporta exclude; ambos se aplican.
func TestLoadConfig_ChainFieldMerge(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.linelens]\ndefault = { maxLines = 10 }\n"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"linelens":{"exclude":["legacy"]}}`), 0644)

	cfg, err := loadConfig(filepath.Join(dir, "linelens.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Default.MaxLines != 10 {
		t.Errorf("esperado maxLines=10 de pyproject, got %d", cfg.Default.MaxLines)
	}
	if !contieneExclude(cfg.Exclude, "legacy") {
		t.Errorf("esperado exclude con 'legacy' de package.json, got %v", cfg.Exclude)
	}
}
