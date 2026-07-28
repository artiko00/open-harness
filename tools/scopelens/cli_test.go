package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// captura ejecuta fn con os.Stdout redirigido y devuelve lo escrito.
func captura(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 1<<16)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func TestRunVersion(t *testing.T) {
	var code int
	out := captura(t, func() { code = run([]string{"version"}) })
	if code != 0 || !strings.Contains(out, "scopelens "+version) {
		t.Fatalf("version: code=%d out=%q", code, out)
	}
}

func TestRunHelp(t *testing.T) {
	if code := runSilent([]string{"help"}); code != 0 {
		t.Fatalf("help code=%d", code)
	}
	if code := runSilent([]string{"--help"}); code != 0 {
		t.Fatalf("--help code=%d", code)
	}
}

func TestRunSinArgs(t *testing.T) {
	if code := runSilent(nil); code != 2 {
		t.Fatalf("sin args code=%d, want 2", code)
	}
}

func TestRunComandoDesconocido(t *testing.T) {
	if code := runSilent([]string{"foo"}); code != 2 {
		t.Fatalf("comando desconocido code=%d, want 2", code)
	}
}

func runSilent(args []string) int {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	code := run(args)
	w.Close()
	os.Stdout = old
	return code
}

func TestRunInit(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "scopelens.json")
	code := runSilent([]string{"init", "--output", out})
	if code != 0 {
		t.Fatalf("init code=%d", code)
	}
	data, err := os.ReadFile(out)
	if err != nil || !strings.Contains(string(data), `"maxFiles": 15`) {
		t.Fatalf("init no generó config válida: %v %s", err, data)
	}
}

func TestRunInitYaExiste(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "scopelens.json")
	os.WriteFile(out, []byte("{}"), 0o644)
	if code := runSilent([]string{"init", "--output", out}); code != 2 {
		t.Fatalf("init existente code=%d, want 2", code)
	}
}

func TestRunInitFlagInvalido(t *testing.T) {
	if code := runSilent([]string{"init", "--nope"}); code != 2 {
		t.Fatalf("init flag inválido code=%d, want 2", code)
	}
}

func TestRunInitErrorEscritura(t *testing.T) {
	out := filepath.Join(t.TempDir(), "noexiste", "scopelens.json")
	if code := runSilent([]string{"init", "--output", out}); code != 2 {
		t.Fatalf("init error escritura code=%d, want 2", code)
	}
}

func TestGitCmdError(t *testing.T) {
	e := &gitCmdError{stderr: "fatal: boom", err: exec.ErrNotFound}
	if e.Error() != "fatal: boom" {
		t.Errorf("Error con stderr = %q", e.Error())
	}
	if e.Unwrap() != exec.ErrNotFound {
		t.Errorf("Unwrap = %v", e.Unwrap())
	}
	e2 := &gitCmdError{err: exec.ErrNotFound}
	if e2.Error() != exec.ErrNotFound.Error() {
		t.Errorf("Error sin stderr = %q", e2.Error())
	}
}

func TestDefaultConfigJSON(t *testing.T) {
	s := defaultConfigJSON()
	for _, want := range []string{`"maxFiles": 15`, "pnpm-lock.yaml", "zz_generated*.go"} {
		if !strings.Contains(s, want) {
			t.Errorf("defaultConfigJSON sin %q", want)
		}
	}
}
