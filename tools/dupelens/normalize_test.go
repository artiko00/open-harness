package main

import (
	"strings"
	"testing"
)

// Un run de tokens idénticos (p. ej. keywords.go tras stripear strings deja
// una tira de "true") no aporta señal estructural: todas sus ventanas son
// monótonas y no compiten en la detección.
func TestAddFile_monotoneRunOnlyYieldsMonotoneWindows(t *testing.T) {
	var files []fileData
	v := newVocab()
	addFile(&files, v, "x.go", strings.Repeat("aa ", 60), scanOpts{windowSize: 25})
	if len(files) != 1 {
		t.Fatalf("se esperaba 1 archivo registrado, got %d", len(files))
	}
	for _, w := range windowsOf(v, files[0], 25, false) {
		if !w.mono {
			t.Fatalf("ventana %d no marcada como monótona", w.start)
		}
	}
}

func TestAddFile_shorterThanWindowIsNotRegistered(t *testing.T) {
	var files []fileData
	addFile(&files, newVocab(), "x.go", "alpha beta", scanOpts{windowSize: 25})
	if len(files) != 0 {
		t.Errorf("un archivo con menos tokens que la ventana no debe registrarse; got %d", len(files))
	}
}

func TestNormalizeValue(t *testing.T) {
	cases := map[string]string{
		"func":  "func", // keyword se preserva
		"42":    "NUM",  // literal numérico
		"3.14":  "NUM",  // numérico con punto
		"foo":   "ID",   // identificador
		"foo_1": "ID",   // identificador con dígito y guion bajo
		"a+b":   "a+b",  // operador glued: ni keyword ni número ni id → literal
	}
	for in, want := range cases {
		if got := normalizeValue(in); got != want {
			t.Errorf("normalizeValue(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestIsIdentifier(t *testing.T) {
	if !isIdentifier("abc1") {
		t.Error("abc1 debe ser identificador")
	}
	if isIdentifier("1abc") {
		t.Error("1abc no debe ser identificador (empieza con dígito)")
	}
	if isIdentifier("a-b") {
		t.Error("a-b no debe ser identificador (guion)")
	}
	if isIdentifier("") {
		t.Error("vacío no debe ser identificador")
	}
}

func TestDeriveWindow(t *testing.T) {
	if got := deriveWindow(0); got != defaultWindow {
		t.Errorf("deriveWindow(0) = %d; want %d", got, defaultWindow)
	}
	if got := deriveWindow(-3); got != defaultWindow {
		t.Errorf("deriveWindow(-3) = %d; want %d", got, defaultWindow)
	}
	if got := deriveWindow(10); got != 10 {
		t.Errorf("deriveWindow(10) = %d; want 10", got)
	}
}
