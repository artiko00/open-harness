package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// checkWith ejecuta runCheck con un gitRunner fake y stdout capturado.
func checkWith(t *testing.T, f *fakeGit, args ...string) (int, string) {
	t.Helper()
	old := newRunner
	newRunner = func(string) gitRunner { return f.run }
	defer func() { newRunner = old }()

	var code int
	out := captura(t, func() { code = runCheck(args) })
	return code, out
}

func manyFiles(n int, ext string) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("src/f")
		b.WriteByte(byte('a' + i))
		b.WriteString(ext)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestCheckParseError(t *testing.T) {
	if code, _ := checkWith(t, fullRepo("", ""), "--nope"); code != 2 {
		t.Fatalf("flag desconocido code=%d, want 2", code)
	}
}

func TestCheckMaxFilesNegativo(t *testing.T) {
	if code, _ := checkWith(t, fullRepo("", ""), "--max-files", "-1"); code != 2 {
		t.Fatalf("max-files<0 code=%d, want 2", code)
	}
}

func TestCheckDirInaccesible(t *testing.T) {
	code, _ := checkWith(t, fullRepo("", ""), "--dir", filepath.Join(t.TempDir(), "noexiste"))
	if code != 2 {
		t.Fatalf("dir inaccesible code=%d, want 2", code)
	}
}

func TestCheckConfigInvalida(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{bad}`)
	if code, _ := checkWith(t, fullRepo("", ""), "--dir", dir); code != 2 {
		t.Fatalf("config inválida code=%d, want 2", code)
	}
}

func TestCheckMeasureError(t *testing.T) {
	f := &fakeGit{shallow: &resp{out: "true\n"}}
	if code, _ := checkWith(t, f, "--dir", t.TempDir()); code != 2 {
		t.Fatalf("shallow code=%d, want 2", code)
	}
}

func TestCheckDentroDelPresupuesto(t *testing.T) {
	f := fullRepo("", manyFiles(3, ".go"))
	code, out := checkWith(t, f, "--dir", t.TempDir(), "--fail", "--no-color")
	if code != 0 {
		t.Fatalf("dentro code=%d, want 0", code)
	}
	if !strings.Contains(out, "OK: 3 files (max 15)") {
		t.Fatalf("salida OK esperada:\n%s", out)
	}
}

func TestCheckExcedeConFail(t *testing.T) {
	f := fullRepo("", manyFiles(18, ".go"))
	code, out := checkWith(t, f, "--dir", t.TempDir(), "--fail", "--no-color")
	if code != 1 {
		t.Fatalf("excede con --fail code=%d, want 1", code)
	}
	if !strings.Contains(out, "FAIL: 18 files (max 15)") {
		t.Fatalf("salida FAIL esperada:\n%s", out)
	}
}

func TestCheckExcedeSinFail(t *testing.T) {
	f := fullRepo("", manyFiles(18, ".go"))
	code, out := checkWith(t, f, "--dir", t.TempDir(), "--no-color")
	if code != 0 {
		t.Fatalf("sin --fail code=%d, want 0", code)
	}
	if !strings.Contains(out, "FAIL: 18 files") {
		t.Fatalf("debe reportar FAIL pero exit 0:\n%s", out)
	}
}

func TestCheckMaxFilesOverride(t *testing.T) {
	f := fullRepo("", manyFiles(3, ".go"))
	code, out := checkWith(t, f, "--dir", t.TempDir(), "--max-files", "2", "--fail", "--no-color")
	if code != 1 {
		t.Fatalf("override max-files=2 con 3 files code=%d, want 1", code)
	}
	if !strings.Contains(out, "max 2") {
		t.Fatalf("límite override no aplicado:\n%s", out)
	}
}

func TestCheckExcludeTests(t *testing.T) {
	// 10 source + 8 test, maxFiles 15, --exclude-tests → contable 10 → exit 0.
	f := fullRepo(manyFiles(8, "_test.go"), manyFiles(10, ".go"))
	code, out := checkWith(t, f, "--dir", t.TempDir(), "--exclude-tests", "--fail", "--no-color")
	if code != 0 {
		t.Fatalf("exclude-tests code=%d, want 0", code)
	}
	if !strings.Contains(out, "OK: 10 files") {
		t.Fatalf("contable con exclude-tests:\n%s", out)
	}
}

func TestCheckExcludeTestsCuentan(t *testing.T) {
	f := fullRepo(manyFiles(8, "_test.go"), manyFiles(10, ".go"))
	code, _ := checkWith(t, f, "--dir", t.TempDir(), "--fail", "--no-color")
	if code != 1 {
		t.Fatalf("sin exclude-tests 18 contables code=%d, want 1", code)
	}
}

func TestCheckConfigDir(t *testing.T) {
	// scopelens.json con maxFiles 2 aplica sin flag.
	dir := t.TempDir()
	writeFile(t, dir, "scopelens.json", `{"maxFiles":2}`)
	f := fullRepo("", manyFiles(3, ".go"))
	code, out := checkWith(t, f, "--dir", dir, "--fail", "--no-color")
	if code != 1 || !strings.Contains(out, "max 2") {
		t.Fatalf("config maxFiles 2 code=%d out=%s", code, out)
	}
}
