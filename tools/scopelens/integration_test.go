package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitOrSkip(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git no está en PATH")
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// nuevoRepo arma un repo con main (1 commit) y una rama feature con un archivo
// committeado, un rename, un borrado y un archivo staged con nombre no ASCII.
func nuevoRepo(t *testing.T) string {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "checkout", "-q", "-b", "main")
	git(t, dir, "config", "core.quotepath", "true")
	writeFile(t, dir, "base.txt", "base\n")
	writeFile(t, dir, "viejo.py", "x\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-qm", "init")
	git(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "nuevo.go", "package a\n")
	git(t, dir, "add", "nuevo.go")
	git(t, dir, "commit", "-qm", "add")
	git(t, dir, "mv", "viejo.py", "renombrado.py")
	os.Remove(filepath.Join(dir, "base.txt"))
	git(t, dir, "rm", "-q", "base.txt")
	writeFile(t, dir, "café.ts", "x\n")
	git(t, dir, "add", ".")
	return dir
}

func TestIntegracionRepoReal(t *testing.T) {
	gitOrSkip(t)
	dir := nuevoRepo(t)
	var code int
	out := captura(t, func() {
		code = runCheck([]string{"--dir", dir, "--base", "main", "--no-color"})
	})
	if code != 0 {
		t.Fatalf("code=%d\n%s", code, out)
	}
	for _, want := range []string{"feature vs main", "café.ts", "renombrado.py", "base.txt", "nuevo.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("falta %q en:\n%s", want, out)
		}
	}
}

func TestIntegracionSinCommits(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	writeFile(t, dir, "a.go", "package a\n")
	writeFile(t, dir, "b.go", "package b\n")
	git(t, dir, "add", ".")
	var code int
	out := captura(t, func() { code = runCheck([]string{"--dir", dir, "--fail", "--no-color"}) })
	if code != 0 || !strings.Contains(out, "sólo staged") {
		t.Fatalf("sin commits code=%d\n%s", code, out)
	}
}

func TestIntegracionNoRepo(t *testing.T) {
	gitOrSkip(t)
	// run(["check"]) cubre el ruteo del subcomando con git real fuera de un repo.
	code := runSilent([]string{"check", "--dir", t.TempDir(), "--fail"})
	if code != 2 {
		t.Fatalf("no repo code=%d, want 2", code)
	}
}

func TestIntegracionContextoCancelado(t *testing.T) {
	gitOrSkip(t)
	dir := nuevoRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newGitRunner(dir)(ctx, "rev-parse", "HEAD"); err == nil {
		t.Fatal("esperaba error por contexto cancelado")
	}
}

func TestMainVersion(t *testing.T) {
	oldExit, oldArgs := osExit, os.Args
	defer func() { osExit, os.Args = oldExit, oldArgs }()
	var code int
	osExit = func(c int) { code = c }
	os.Args = []string{"scopelens", "version"}
	captura(t, func() { main() })
	if code != 0 {
		t.Fatalf("main version code=%d", code)
	}
}
