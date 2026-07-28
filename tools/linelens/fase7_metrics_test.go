package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artiko00/open-harness/tools/_shared/pathmatch"
)

func TestPhysicalLines(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"a\nb\n", 2},
		{"a\nb", 2},   // última línea sin salto también cuenta
		{"package main\n", 1},
	}
	for _, c := range cases {
		if got := physicalLines([]byte(c.in)); got != c.want {
			t.Errorf("physicalLines(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestNonBlankLines(t *testing.T) {
	got := nonBlankLines("uno\n\n  \ndos\n")
	if got != 2 {
		t.Errorf("nonBlankLines = %d, want 2", got)
	}
}

func TestMaxLineLen(t *testing.T) {
	// Sin salto final: la rama post-bucle debe considerar la última línea.
	if got := maxLineLen([]byte("ab\nabcd")); got != 4 {
		t.Errorf("maxLineLen sin salto final = %d, want 4", got)
	}
	if got := maxLineLen([]byte("abcde\n")); got != 5 {
		t.Errorf("maxLineLen con salto final = %d, want 5", got)
	}
}

func TestMaxNesting(t *testing.T) {
	if got := maxNesting("a { b { c } } d"); got != 2 {
		t.Errorf("maxNesting balanceado = %d, want 2", got)
	}
	// Llaves de cierre de más: el guard depth>0 evita ir a negativo.
	if got := maxNesting("} } { x }"); got != 1 {
		t.Errorf("maxNesting con cierres extra = %d, want 1", got)
	}
}

func TestCountFile_Metricas(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.go")
	os.WriteFile(path, []byte("package p\n// c\nfunc F() {\n\tif x { y }\n}\n"), 0644)

	code, physical, nesting, err := countFile(path, ".go")
	if err != nil {
		t.Fatal(err)
	}
	if physical != 5 {
		t.Errorf("physical = %d, want 5", physical)
	}
	if code != 4 { // se descuenta la línea de comentario
		t.Errorf("code = %d, want 4", code)
	}
	if nesting != 2 {
		t.Errorf("nesting = %d, want 2", nesting)
	}
}

func TestMotivoLectura_LineTooLong(t *testing.T) {
	if got := motivoLectura(errLineTooLong); got != pathmatch.ReasonLineTooLong {
		t.Errorf("motivoLectura(errLineTooLong) = %q, want %q", got, pathmatch.ReasonLineTooLong)
	}
	if got := motivoLectura(os.ErrPermission); got != pathmatch.ReasonReadError {
		t.Errorf("motivoLectura(otro) = %q, want %q", got, pathmatch.ReasonReadError)
	}
}

func TestCountFile_LineaMuyLarga(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "huge.go")
	os.WriteFile(path, []byte(strings.Repeat("a", pathmatch.BufferLimit+1)+"\n"), 0644)

	if _, _, _, err := countFile(path, ".go"); err != errLineTooLong {
		t.Errorf("esperaba errLineTooLong, got %v", err)
	}
}
